package wordTranslator

import (
	"strings"
)

type defaultStrategy struct{}

func init() {
	RegisterStrategy("default", &defaultStrategy{})
}

func (s *defaultStrategy) process(tr *TranslateResult, raw []interface{}) {
	tr.IsSimpleWord = len(strings.Fields(tr.SourceWord)) == 1

	if first, ok := safeGetArray(raw, 0); ok {
		if block, ok := safeGetArray(first, 1); ok {
			if len(block) >= 4 {
				if t, ok := safeGetString(block, 3); ok {
					tr.Transcription = t
				}
			} else if len(block) >= 3 {
				if t, ok := safeGetString(block, 2); ok {
					tr.Transcription = t
				}
			}
		}
	}

	if units, ok := safeGetArray(raw, 12); ok && len(units) > 0 {
		if firstUnit, ok := safeGetArray(units, 0); ok && len(firstUnit) > 2 {
			if typeIndicator, ok := safeGetString(firstUnit, 0); ok {
				switch typeIndicator {
				case "":
					if forms, ok := safeGetArray(firstUnit, 1); ok && len(forms) > 0 {
						if infPair, ok := safeGetArray(forms, 0); ok && len(infPair) > 0 {
							if inf, ok := safeGetString(infPair, 0); ok {
								tr.Infinitive = inf
							}
						}
					}
				case "verb":
					if inf, ok := safeGetString(firstUnit, 2); ok {
						tr.Infinitive = inf
					}
				default:
				}
			}
		}
	}

	if tr.Infinitive == "" {
		if transUnits, ok := safeGetArray(raw, 5); ok && len(transUnits) > 0 {
			if firstUnit, ok := safeGetArray(transUnits, 0); ok && len(firstUnit) > 0 {
				if inf, ok := safeGetString(firstUnit, 0); ok {
					tr.Infinitive = inf
				}
			}
		}
	}
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
	}

	tr.Examples = nil

	if posRaw, ok := safeGetArray(raw, 12); ok {

	GatherPOS:
		for _, p := range posRaw {
			entry, ok := p.([]interface{})
			if !ok || len(entry) < 2 {
				continue
			}
			posStr, _ := entry[0].(string)
			tr.PartOfSpeech = posStr
			switch posStr {
			case "noun":
				if senses, ok := safeGetArray(entry, 1); ok {
					for _, sRaw := range senses {
						sense, ok := sRaw.([]interface{})
						if ok && len(sense) > 2 {
							if exText, ok := sense[2].(string); ok {
								tr.Examples = append(tr.Examples, exText)
							}
						}
						if len(tr.Examples) >= 5 {
							break GatherPOS
						}
					}
				}
			case "verb", "adverb":
				if exOuter, ok := safeGetArray(raw, 13); ok {
					for _, group := range exOuter {
						if list, ok := group.([]interface{}); ok {
							for _, item := range list {
								if ent, ok := item.([]interface{}); ok && len(ent) > 0 {
									if txt, ok := ent[0].(string); ok {
										tr.Examples = append(tr.Examples, txt)
									}
								}
								if len(tr.Examples) >= 5 {
									break GatherPOS
								}
							}
						}
					}
				}
			default:
				if exOuter, ok := safeGetArray(raw, 13); ok {
					for _, group := range exOuter {
						if list, ok := group.([]interface{}); ok {
							for _, item := range list {
								if ent, ok := item.([]interface{}); ok && len(ent) > 0 {
									if txt, ok := ent[0].(string); ok {
										tr.Examples = append(tr.Examples, txt)
									}
								}
								if len(tr.Examples) >= 5 {
									break GatherPOS
								}
							}
						}
					}
				}
				break
			}
		}

		tr.Translations = SliceUnique(tr.Translations)
		tr.Examples = SliceUnique(tr.Examples)

		conf, ok := safeGetFloat(raw, 6)
		isCorrect := ok && conf > 0.75

		if len(tr.Translations) > 0 || (isCorrect) {
			tr.IsValid = true
		}
	}
}

func (s *defaultStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	s.process(tr, raw)
}
