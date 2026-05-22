package cards

import (
	"context"

	"encore.dev/beta/errs"
)

//encore:api public method=GET path=/banks
func ListBanks(ctx context.Context) (*ListBanksResponse, error) {
	banks, err := queryBanks(ctx)
	if err != nil {
		return nil, err
	}
	return &ListBanksResponse{Banks: banks}, nil
}

//encore:api public method=POST path=/cards
func CreateCard(ctx context.Context, p *CreateCardParams) (*Card, error) {
	exists, err := bankExists(ctx, p.BankID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.B().Code(errs.NotFound).Msg("bank not found").Err()
	}

	return insertCard(ctx, p)
}

//encore:api public method=GET path=/cards
func ListCards(ctx context.Context) (*ListCardsResponse, error) {
	cards, err := queryCards(ctx)
	if err != nil {
		return nil, err
	}
	return &ListCardsResponse{Cards: cards}, nil
}

//encore:api public method=GET path=/cards/:id
func GetCard(ctx context.Context, id string) (*Card, error) {
	return getCardByID(ctx, id)
}
