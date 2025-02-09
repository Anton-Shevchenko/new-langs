package word_repository

import (
	"fmt"
	"gorm.io/gorm"
	"langs/internal/model"
	"math/rand"
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
