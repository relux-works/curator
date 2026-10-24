package projectcontext

// inspectionStatus describes source isolation independently of the item admission
// switch. Inspection itself grants no authority to deliver project items.
func inspectionStatus(registration Registration, items []Discovery) Inspection {
	out := Inspection{
		Registration: registration,
		Mode:         registration.Mode,
		Admission:    "off",
		Memory:       "off",
		Rendering:    AppendAfterProfile,
		Candidates:   items,
	}
	if registration.Mode == "native-project" {
		out.Notice = "project sources not isolated"
	}
	return out
}
