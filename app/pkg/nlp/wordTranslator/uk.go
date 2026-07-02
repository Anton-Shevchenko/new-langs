package wordTranslator

type ukrainianStrategy struct{}

func init() {
	RegisterStrategy("uk", &ukrainianStrategy{})
}

func (s *ukrainianStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	if transUnits, ok := safeGetArray(raw, 5); ok && len(transUnits) > 0 {
		for _, u := range transUnits {
			if unit, ok := u.([]interface{}); ok && len(unit) > 2 {
				if forms, ok := unit[2].([]interface{}); ok {
					for _, f := range forms {
						if fa, ok := f.([]interface{}); ok && len(fa) > 0 {
							if w, ok := fa[0].(string); ok {
								tr.Translations = append(tr.Translations, w)
							}
						}
					}
				}
			}
		}
		tr.Translations = SliceUnique(tr.Translations)
	}

	if conf, ok := safeGetFloat(raw, 6); ok && conf < 0.75 && len(tr.Translations) == 0 {
		tr.IsValid = false
		return
	}

	tr.IsValid = true
}
