package wordTranslator

type englishStrategy struct {
	defaultStrategy
}

func init() {
	RegisterStrategy("en", &englishStrategy{})
}

func (s *englishStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	s.process(tr, raw)
	if first, ok := safeGetArray(raw, 0); ok {
		if block, ok := safeGetArray(first, 1); ok && len(block) >= 3 {
			if t, ok := safeGetString(block, 2); ok {
				tr.Transcription = t
			}
		}
	}
}
