package repository

import (
	"bpl/client"
	"bpl/config"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type StashChangeRepository struct {
	DB *gorm.DB
}
type StashChange struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	StashID      string `gorm:"not null"`
	NextChangeID string `gorm:"not null"`
	EventID      int    `gorm:"index;not null;references events(id)"`
	Timestamp    time.Time
}

func NewStashChangeRepository() *StashChangeRepository {
	return &StashChangeRepository{DB: config.DatabaseConnection()}
}

func (r *StashChangeRepository) SaveStashChangesConditionally(publicStashChanges []client.PublicStashChange, message config.StashChangeMessage, eventId int, sendFunc func([]byte) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		dbStashChanges := make([]StashChange, 0)
		for _, stashChange := range publicStashChanges {
			dbStashChanges = append(dbStashChanges, StashChange{
				StashID:      stashChange.ID,
				EventID:      eventId,
				NextChangeID: message.NextChangeID,
				Timestamp:    message.Timestamp,
			})
		}

		result := r.DB.CreateInBatches(dbStashChanges, 500)
		if result.Error != nil {
			return result.Error
		}
		idMap := make(map[string]int)
		for _, stashChange := range dbStashChanges {
			idMap[stashChange.StashID] = stashChange.ID
		}
		for i, s := range publicStashChanges {
			publicStashChanges[i].StashChangeID = idMap[s.ID]
		}
		message.Stashes = publicStashChanges

		data, err := json.Marshal(message)

		if err != nil {
			return err
		}
		return sendFunc(data)
	})
}
func (r *StashChangeRepository) GetNextChangeIdForEvent(event *Event) (changeId string, err error) {
	query := "SELECT DISTINCT next_change_id FROM stash_changes WHERE event_id = ? ORDER BY next_change_id desc LIMIT 1"
	err = r.DB.Raw(query, event.ID).First(&changeId).Error
	return changeId, err
}

func (r *StashChangeRepository) GetCurrentChangeIdForEvent(event *Event) (changeId string, err error) {
	query := "SELECT DISTINCT next_change_id FROM stash_changes WHERE event_id = ? ORDER BY next_change_id desc OFFSET 1 LIMIT 1"
	err = r.DB.Raw(query, event.ID).First(&changeId).Error
	return changeId, err
}
