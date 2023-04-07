package facade

func toSave(entitiesToSave []db.EntityHolder, entityHolder db.EntityHolder) []db.EntityHolder {
	kind, intID := entityHolder.Kind(), entityHolder.IntID()
	for _, e := range entitiesToSave {
		if e.Kind() == kind && e.IntID() == intID {
			return entitiesToSave
		}
	}
	entitiesToSave = append(entitiesToSave, entityHolder)
	return entitiesToSave
}
