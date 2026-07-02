package word_repository

import (
	"fmt"
	"gorm.io/gorm"
	"langs/internal/domain"
	"math/rand"
	"strings"
	"time"
)

type WordRepository struct {
	db *gorm.DB
}

func NewWordRepository(db *gorm.DB) *WordRepository {
	return &WordRepository{
		db: db,
	}
}

func (r *WordRepository) Create(word *model.Word) error {
	return r.db.Create(&word).Error
}

func (r *WordRepository) Save(word *model.Word) error {
	return r.db.Where("id = ?", word.ID).Save(&word).Error
}

func (r *WordRepository) AllByChatId(chatId int64, limit, offset int) ([]*model.Word, error) {
	var words []*model.Word

	if limit > 100 {
		limit = 100
	}

	return words, r.db.
		Where("chat_id = ?", chatId).
		Order("id desc").
		Limit(limit).
		Offset(offset).
		Find(&words).Error
}

func (r *WordRepository) SearchByChatId(chatId int64, query string, limit int) ([]*model.Word, error) {
	var words []*model.Word

	pattern := "%" + strings.ToLower(query) + "%"

	return words, r.db.
		Where(
			"chat_id = ? AND (LOWER(value) LIKE ? OR LOWER(translation) LIKE ?)",
			chatId, pattern, pattern,
		).
		Order("id desc").
		Limit(limit).
		Find(&words).Error
}

// GetAllByChatId returns every word for the chat, newest first. Unlike
// AllByChatId it is not paginated and is intended for exports.
func (r *WordRepository) GetAllByChatId(chatId int64) ([]*model.Word, error) {
	var words []*model.Word
	return words, r.db.
		Where("chat_id = ?", chatId).
		Order("id desc").
		Find(&words).Error
}

// AllByChatIdAndLangPair returns every word for the chat whose languages match
// the given pair in either direction (lang1->lang2 or lang2->lang1).
func (r *WordRepository) AllByChatIdAndLangPair(chatId int64, lang1, lang2 string) ([]*model.Word, error) {
	var words []*model.Word
	return words, r.db.
		Where(
			"chat_id = ? AND ((value_lang = ? AND translation_lang = ?) OR (value_lang = ? AND translation_lang = ?))",
			chatId, lang1, lang2, lang2, lang1,
		).
		Order("id desc").
		Find(&words).Error
}

// GetLangPairsByChatId returns the distinct, direction-independent language
// pairs the user has saved words for, with a total count per pair.
func (r *WordRepository) GetLangPairsByChatId(chatId int64) ([]model.LangPair, error) {
	type row struct {
		ValueLang       string
		TranslationLang string
		Count           int
	}

	var rows []row
	err := r.db.
		Model(&model.Word{}).
		Select("value_lang, translation_lang, COUNT(*) as count").
		Where("chat_id = ?", chatId).
		Group("value_lang, translation_lang").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	merged := map[string]*model.LangPair{}
	var order []string
	for _, rw := range rows {
		a, b := rw.ValueLang, rw.TranslationLang
		if a == "" || b == "" || a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		key := a + "_" + b
		if merged[key] == nil {
			merged[key] = &model.LangPair{Lang1: a, Lang2: b}
			order = append(order, key)
		}
		merged[key].Count += rw.Count
	}

	pairs := make([]model.LangPair, 0, len(order))
	for _, key := range order {
		pairs = append(pairs, *merged[key])
	}
	return pairs, nil
}

func (r *WordRepository) GetCountByChatId(chatId int64) int64 {
	var count int64
	r.db.Model(&model.Word{}).Where("chat_id = ?", chatId).Count(&count)

	return count
}

func (r *WordRepository) First(id int64) (*model.Word, error) {
	var word model.Word
	err := r.db.Where("id = ?", id).First(&word).Error

	if err != nil {
		return nil, err
	}

	return &word, nil
}

func (r *WordRepository) Delete(id int64) error {
	return r.db.Delete(&model.Word{}, id).Error
}

func (r *WordRepository) GetRandomWordByChatIdAndRateLimit(chatId int64, rate uint16) (*model.Word, error) {
	var words []model.Word
	if err := r.db.
		Where("chat_id = ? AND rate < ?", chatId, rate).
		Order("RANDOM()").
		Limit(5).
		Find(&words).
		Error; err != nil {
		return nil, err
	}

	if len(words) == 0 {
		return nil, fmt.Errorf("no words found")
	}

	rGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomIndex := rGen.Intn(len(words))
	return &words[randomIndex], nil
}

func (r *WordRepository) GetRandomTranslationsByChatId(
	chatId int64,
	exception string,
	lang string,
	limit int,
) ([]string, error) {
	var words []string

	return words, r.db.
		Table("words").
		Select(
			"DISTINCT ON (word)  CASE WHEN value_lang = ? THEN value WHEN translation_lang = ? THEN translation END AS word",
			lang, lang,
		).
		Where(
			"chat_id = ? AND value != ? AND translation != ? AND (value_lang = ? OR translation_lang = ?)",
			chatId, exception, exception, lang, lang,
		).
		Order("word, RANDOM()").
		Limit(limit).
		Pluck("word", &words).Error
}

func (r *WordRepository) CheckSimilarWord(word *model.Word) (*model.Word, error) {
	var savedWord model.Word

	return &savedWord, r.db.Where(
		"chat_id = ? AND ("+
			"(value = ? AND value_lang = ?) OR (translation = ? AND translation_lang = ?) "+
			"OR "+
			"(value = ? AND value_lang = ?) OR (translation = ? AND translation_lang = ?)"+
			")",
		word.ChatId,
		word.Value, word.ValueLang, word.Translation, word.TranslationLang,
		word.Translation, word.TranslationLang, word.Value, word.ValueLang,
	).Find(&savedWord).Error
}
