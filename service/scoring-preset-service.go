package service

import (
	"bpl/repository"
)

type ScoringPresetsService struct {
	scoring_preset_repository *repository.ScoringPresetRepository
	objective_repository      *repository.ObjectiveRepository
}

func NewScoringPresetsService() *ScoringPresetsService {
	return &ScoringPresetsService{
		scoring_preset_repository: repository.NewScoringPresetRepository(),
		objective_repository:      repository.NewObjectiveRepository(),
	}
}

func (s *ScoringPresetsService) SavePreset(preset *repository.ScoringPreset) (*repository.ScoringPreset, error) {
	return s.scoring_preset_repository.SavePreset(preset)
}

func (s *ScoringPresetsService) GetPresetById(presetId int) (*repository.ScoringPreset, error) {
	return s.scoring_preset_repository.GetPresetById(presetId)
}

func (s *ScoringPresetsService) GetPresetsForEvent(eventId int) ([]*repository.ScoringPreset, error) {
	return s.scoring_preset_repository.GetPresetsForEvent(eventId)
}

func (s *ScoringPresetsService) DeletePreset(presetId int) error {
	err := s.objective_repository.RemoveScoringId(presetId)
	if err != nil {
		return err
	}
	return s.scoring_preset_repository.DeletePreset(presetId)
}
