package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/store"
	"gorm.io/gorm"
)

// CreateEntry optionally deduplicates creation. A retry reuses the original input and key.
func (s *Service) CreateEntry(in EntryInput, key string) (*model.MealEntry, error) {
	if key == "" {
		return s.SaveEntry(0, in)
	}
	// Hash supplied fields before resolving defaults such as today's date.
	// Field[T] also retains omission/null; JSON field order and number formatting do not matter.
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	var result *model.MealEntry
	err = s.transaction(func(tx *Service) error {
		request, err := tx.store.MealEntryRequest(key)
		if err == nil {
			result, err = replayEntry(request, hash)
			return err
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := in.validateNulls(); err != nil {
			return err
		}
		result, err = tx.saveEntry(0, in)
		if err != nil {
			return err
		}
		return tx.store.CreateMealEntryRequest(&model.MealEntryRequest{
			Key: key, RequestHash: hash, Response: *result,
		})
	})
	if errors.Is(err, store.ErrMealEntryRequestExists) {
		// A concurrent winner has committed. The losing entry was rolled back;
		// read outside its REPEATABLE READ transaction to see the winner's result.
		request, readErr := s.store.MealEntryRequest(key)
		if readErr != nil {
			return nil, readErr
		}
		return replayEntry(request, hash)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func replayEntry(request *model.MealEntryRequest, hash string) (*model.MealEntry, error) {
	if request.RequestHash != hash {
		return nil, &Error{
			Code: "idempotency_conflict", Message: "请求标识已用于不同的饮食记录内容",
			Fields: map[string]string{"Idempotency-Key": "同一标识重试时须使用相同内容；新的操作请使用新标识"},
		}
	}
	return &request.Response, nil
}
