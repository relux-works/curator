from pathlib import Path
from types import SimpleNamespace
import sqlite3
import pytest
import csk_registry.store as module
from test_registry import _health_client, _body


def test_review_concurrent_append_between_verifier_reads(tmp_path, monkeypatch):
    client, store, key, _, _ = _health_client(tmp_path)
    record = key.sign_record(_body())
    store.append(record, created_at='2026-07-13T02:00:00Z')
    original = store._load_frontier
    injected = False
    def interleaved(conn=None):
        nonlocal injected
        if conn is not None and not injected:
            injected = True
            store.append(record, created_at='2026-07-13T02:00:01Z')
        return original(conn)
    monkeypatch.setattr(store, '_load_frontier', interleaved)
    verdict = store.refresh_health_verdict()
    assert injected
    assert store.integrity_errors() == []
    assert verdict.ready, verdict.error
    assert client.get('/health').status_code == 200


def test_review_staleness_must_latch_until_restart(tmp_path, monkeypatch):
    client, store, key, _, _ = _health_client(tmp_path, health_verify_interval=10)
    base = store._health_last_refresh_monotonic
    monkeypatch.setattr(module, 'time', SimpleNamespace(monotonic=lambda: base + 20.001))
    assert client.get('/health').status_code == 503
    with pytest.raises(module.StoreIntegrityError):
        store.append(key.sign_record(_body()), created_at='2026-07-13T02:00:00Z')
    assert not store.refresh_health_verdict().ready, 'staleness recovered without restart'


def test_review_append_does_not_attest_corrupted_frontier(tmp_path):
    client, store, key, _, _ = _health_client(tmp_path)
    record = key.sign_record(_body())
    for _ in range(3):
        store.append(record, created_at='2026-07-13T02:00:00Z')
    with sqlite3.connect(store.path) as conn:
        conn.execute('UPDATE merkle_frontier SET tail = ? WHERE level=0', ('["'+'f'*64+'","'+'e'*64+'"]',))
    with pytest.raises(module.StoreIntegrityError):
        store.append(record, created_at='2026-07-13T02:00:01Z')


def test_review_rollback_does_not_advance_cached_head(tmp_path):
    client, store, key, _, _ = _health_client(tmp_path)
    record = key.sign_record(_body())
    with sqlite3.connect(store.path) as conn:
        conn.execute("CREATE TRIGGER refuse_import BEFORE INSERT ON imported_records BEGIN SELECT RAISE(ABORT, 'injected write failure'); END")
    with pytest.raises(sqlite3.IntegrityError):
        store.append_imports([('fingerprint',record)], created_at='2026-07-13T02:00:00Z')
    assert store.head()[0] == 0
    verdict = store.health_verdict()
    assert verdict.verified_log_size == 0, f'rolled back store reports {verdict}'


def test_review_frontier_corruption_creates_wrong_boundary(tmp_path):
    client, store, key, _, _ = _health_client(tmp_path)
    record = key.sign_record(_body())
    for _ in range(3):
        store.append(record, created_at='2026-07-13T02:00:00Z')
    with sqlite3.connect(store.path) as conn:
        conn.execute('UPDATE merkle_frontier SET tail = ? WHERE level=0', ('["'+'f'*64+'","'+'e'*64+'"]',))
    store.append(record, created_at='2026-07-13T02:00:01Z')
    assert client.get('/health').status_code == 200
    expected = module._merkle_root([entry.entry_hash for entry in store.log_entries()])
    assert store.snapshot_boundary().merkle_root == expected, 'append committed a corrupt root while health stayed 200'
