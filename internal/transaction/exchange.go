package transaction

// Exchange targets keep the published preimage in place while moving the
// complete desired sibling to BackupPath. A single exchange then publishes
// the desired tree and retains the preimage at BackupPath. KindExchange makes
// these two layouts distinguishable from legacy backup/install journals.
func (engine *Engine) commitExchange(journal *Journal, index int) error {
	target := &journal.Targets[index]
	if target.State == StateCommitted {
		current, err := target.digest(target.LivePath)
		if err != nil {
			return err
		}
		if current != target.DesiredDigest {
			return corruptionf("committed exchange target %s/%s changed", target.Class, target.Identifier)
		}
		return nil
	}
	if target.State == StatePending {
		backup, err := target.digest(target.BackupPath)
		if err != nil {
			return err
		}
		if backup == DigestAbsent {
			if err := engine.emit(journal, index, PointBeforeBackup); err != nil {
				return err
			}
			if err := verifyPreimage(target); err != nil {
				return err
			}
			if err := syncTree(target.LivePath); err != nil {
				return err
			}
			if staged, err := target.digest(target.StagedPath); err != nil {
				return err
			} else if staged != target.DesiredDigest {
				return corruptionf("exchange staging changed for %s/%s", target.Class, target.Identifier)
			}
			if err := engine.checkBoundary(journal, index); err != nil {
				return err
			}
			if err := engine.renameNoReplace(target.StagedPath, target.BackupPath); err != nil {
				return err
			}
			if err := engine.emit(journal, index, PointAfterBackup); err != nil {
				return err
			}
		} else if backup != target.DesiredDigest {
			return corruptionf("pending exchange backup changed for %s/%s", target.Class, target.Identifier)
		}
		target.BackupDigest = target.PreimageDigest
		target.State = StateBackedUp
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	if target.State != StateBackedUp {
		return journalf("exchange target %s/%s has invalid commit state", target.Class, target.Identifier)
	}
	current, backup, err := exchangeLayout(target)
	if err != nil {
		return err
	}
	// A crash after the exchange, before the state write, leaves the exact
	// inverse layout. Never exchange that layout a second time on recovery.
	if current == target.PreimageDigest && backup == target.DesiredDigest {
		if err := engine.emit(journal, index, PointBeforeInstall); err != nil {
			return err
		}
		current, backup, err = exchangeLayout(target)
		if err != nil {
			return err
		}
		if current != target.PreimageDigest || backup != target.DesiredDigest {
			return corruptionf("exchange target %s/%s changed before install", target.Class, target.Identifier)
		}
		if err := engine.checkBoundary(journal, index); err != nil {
			return err
		}
		if err := engine.exchange(target.BackupPath, target.LivePath); err != nil {
			return err
		}
		if err := engine.emit(journal, index, PointAfterInstall); err != nil {
			return err
		}
	} else if current != target.DesiredDigest || backup != target.PreimageDigest {
		return corruptionf("exchange target %s/%s contains unknown bytes", target.Class, target.Identifier)
	}
	target.State = StateCommitted
	if err := engine.saveJournal(journal); err != nil {
		return err
	}
	return engine.emit(journal, index, PointTargetCommitted)
}

// exchangeLayout checks the two names involved in the exchange and refuses
// unexpected spare staging or rollback bytes before a live mutation.
func exchangeLayout(target *TargetRecord) (string, string, error) {
	for _, path := range []string{target.StagedPath, target.RollbackPath} {
		if digest, err := target.digest(path); err != nil {
			return "", "", err
		} else if digest != DigestAbsent {
			return "", "", corruptionf("exchange target %s/%s has unexpected sidecar bytes", target.Class, target.Identifier)
		}
	}
	current, err := target.digest(target.LivePath)
	if err != nil {
		return "", "", err
	}
	backup, err := target.digest(target.BackupPath)
	return current, backup, err
}

func (engine *Engine) rollbackExchange(journal *Journal, index int) error {
	target := &journal.Targets[index]
	if target.State == StateRolledBack {
		return nil
	}
	current, err := target.digest(target.LivePath)
	if err != nil {
		return err
	}
	backup, err := target.digest(target.BackupPath)
	if err != nil {
		return err
	}
	if current == target.DesiredDigest && backup == target.PreimageDigest && current != target.PreimageDigest {
		if _, _, err := exchangeLayout(target); err != nil {
			return err
		}
		if err := engine.exchange(target.BackupPath, target.LivePath); err != nil {
			return err
		}
		if err := engine.emit(journal, index, PointAfterRestore); err != nil {
			return err
		}
		current, backup = target.PreimageDigest, target.DesiredDigest
	}
	if current != target.PreimageDigest {
		return corruptionf("rollback refused unknown exchange preimage for %s/%s", target.Class, target.Identifier)
	}
	// The desired backup can exist before installation, after restoration,
	// or after a crash before the restoration state write. Move it to the
	// ordinary rollback name so existing exact-digest cleanup owns it.
	if backup == target.DesiredDigest {
		if _, _, err := exchangeLayout(target); err != nil {
			return err
		}
		if err := engine.renameNoReplace(target.BackupPath, target.RollbackPath); err != nil {
			return err
		}
	} else if backup != DigestAbsent {
		return corruptionf("rollback refused unknown exchange backup for %s/%s", target.Class, target.Identifier)
	}
	target.BackupDigest = target.PreimageDigest
	target.State = StateRolledBack
	if err := engine.saveJournal(journal); err != nil {
		return err
	}
	return engine.emit(journal, index, PointTargetRolledBack)
}

// Exchange mutations use the same acquisition check as ordinary renames, after
// boundary callbacks and layout reads. Rollback checks ownership independently
// of forward admission so the current owner can restore rejected transactions.
func (engine *Engine) exchange(first, second string) error {
	if err := engine.checkOwnership(); err != nil {
		return err
	}
	return durableExchange(first, second)
}
