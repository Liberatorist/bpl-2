package scoring

import (
	"bpl/client"
	"bpl/config"
	"bpl/repository"
	"bpl/service"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type FetchingService struct {
	ctx                context.Context
	event              *repository.Event
	poeClient          *client.PoEClient
	stashChangeService *service.StashChangeService
	stashChannel       chan StashChange
}

func NewFetchingService(ctx context.Context, event *repository.Event, poeClient *client.PoEClient) *FetchingService {
	stashChangeService := service.NewStashChangeService()

	return &FetchingService{
		ctx:                ctx,
		event:              event,
		poeClient:          poeClient,
		stashChangeService: stashChangeService,
		stashChannel:       make(chan StashChange),
	}
}

func (f *FetchingService) FetchStashChanges() error {
	token := os.Getenv("POE_CLIENT_TOKEN")
	if token == "" {
		return fmt.Errorf("POE_CLIENT_TOKEN environment variable not set")
	}
	initialStashChange, err := f.stashChangeService.GetInitialChangeId(f.event)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	changeId := initialStashChange.NextChangeID
	for {
		select {
		case <-f.ctx.Done():
			return nil
		default:
			fmt.Println("Fetching stashes with change id:", changeId)
			response, err := f.poeClient.GetPublicStashes(token, "pc", changeId)
			if err != nil {
				if err.StatusCode == 429 {
					fmt.Println(err.ResponseHeaders)
					retryAfter, err := strconv.Atoi(err.ResponseHeaders.Get("Retry-After"))
					if err != nil {
						fmt.Println(err)
						return fmt.Errorf("failed to parse Retry-After header: %s", err)
					}
					<-time.After((time.Duration(retryAfter) + 1) * time.Second)
				} else {
					fmt.Println(err)
					return fmt.Errorf("failed to fetch public stashes: %s", err.Description)
				}
				continue
			}
			f.stashChannel <- StashChange{ChangeID: changeId, NextChangeID: response.NextChangeID, Stashes: response.Stashes}
			changeId = response.NextChangeID
		}
	}
}

func (f *FetchingService) FilterStashChanges() {
	err := config.CreateTopic(f.event.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	writer, err := config.GetWriter(f.event.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer writer.Close()

	for stashChange := range f.stashChannel {
		select {
		case <-f.ctx.Done():
			return
		default:

			filteredStashChange := StashChange{
				ChangeID:     stashChange.ChangeID,
				NextChangeID: stashChange.NextChangeID,
				Stashes:      make([]client.PublicStashChange, 0),
			}
			now := time.Now()
			stashChanges := make([]*repository.StashChange, 0)
			for _, stash := range stashChange.Stashes {
				intStashChange, err := stashChangeToInt(stashChange.ChangeID)
				if err != nil {
					fmt.Println(err)
					return
				}
				// if stash.League != nil && *stash.League == event.Name {
				stashChanges = append(stashChanges, &repository.StashChange{
					StashID:      stash.ID,
					NextChangeID: stashChange.NextChangeID,
					IntChangeID:  intStashChange,
					EventID:      f.event.ID,
					Timestamp:    now,
				})

				filteredStashChange.Stashes = append(filteredStashChange.Stashes, stash)
				// }
			}
			filteredStashChange.Timestamp = now
			fmt.Printf("Found %d stashes\n", len(filteredStashChange.Stashes))
			data, err := json.Marshal(filteredStashChange)
			if err != nil {
				fmt.Println(err)
				return
			}
			// make sure that stash changes are only saved if the messages are successfully written to kafka
			f.stashChangeService.SaveStashChangesConditionally(stashChanges,
				func() error {
					return writer.WriteMessages(context.Background(),
						kafka.Message{
							Value: data,
						},
					)
				})
		}
	}
}

func FetchLoop(ctx context.Context, event *repository.Event, poeClient *client.PoEClient) {
	fetchingService := NewFetchingService(ctx, event, poeClient)
	go fetchingService.FetchStashChanges()
	go fetchingService.FilterStashChanges()
}
