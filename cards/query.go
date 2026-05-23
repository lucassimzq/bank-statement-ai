package cards

import (
	"context"

	cardsdb "encore.app/cards/db"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

func queryBanks(ctx context.Context) ([]*Bank, error) {
	rows, err := queries.QueryBanks(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query banks")
	}
	banks := make([]*Bank, len(rows))
	for i, r := range rows {
		var logoURL *string
		if r.LogoUrl.Valid {
			logoURL = &r.LogoUrl.String
		}
		banks[i] = &Bank{ID: r.ID.String(), Name: r.Name, Slug: r.Slug, LogoURL: logoURL, CreatedAt: r.CreatedAt}
	}
	return banks, nil
}

func bankExists(ctx context.Context, bankID string) (bool, error) {
	id, err := uuid.Parse(bankID)
	if err != nil {
		return false, errs.B().Code(errs.InvalidArgument).Msg("invalid bank_id").Err()
	}
	exists, err := queries.BankExists(ctx, id)
	if err != nil {
		return false, errs.WrapCode(err, errs.Internal, "check bank exists")
	}
	return exists, nil
}

func insertCard(ctx context.Context, p *CreateCardParams) (*Card, error) {
	bankID, err := uuid.Parse(p.BankID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid bank_id").Err()
	}
	row, err := queries.InsertCard(ctx, cardsdb.InsertCardParams{
		BankID:  bankID,
		Label:   p.Label,
		Purpose: p.Purpose,
		Last4:   p.Last4,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "insert card")
	}
	return &Card{
		ID:        row.ID.String(),
		BankID:    row.BankID.String(),
		Label:     row.Label,
		Purpose:   row.Purpose,
		Last4:     row.Last4,
		CreatedAt: row.CreatedAt,
	}, nil
}

func queryCards(ctx context.Context) ([]*Card, error) {
	rows, err := queries.QueryCards(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query cards")
	}
	cards := make([]*Card, len(rows))
	for i, r := range rows {
		cards[i] = &Card{
			ID:        r.ID.String(),
			BankID:    r.BankID.String(),
			BankName:  r.BankName,
			Label:     r.Label,
			Purpose:   r.Purpose,
			Last4:     r.Last4,
			CreatedAt: r.CreatedAt,
		}
	}
	return cards, nil
}

func getCardByID(ctx context.Context, id string) (*Card, error) {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid card id").Err()
	}
	r, err := queries.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, errs.B().Code(errs.NotFound).Msg("card not found").Err()
	}
	return &Card{
		ID:        r.ID.String(),
		BankID:    r.BankID.String(),
		BankName:  r.BankName,
		Label:     r.Label,
		Purpose:   r.Purpose,
		Last4:     r.Last4,
		CreatedAt: r.CreatedAt,
	}, nil
}
