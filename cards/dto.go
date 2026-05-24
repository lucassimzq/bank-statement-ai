package cards

import "encore.dev/beta/errs"

type ListBanksResponse struct {
	Banks []*Bank `json:"banks"`
}

type CreateCardParams struct {
	BankID  string `json:"bank_id"`
	Label   string `json:"label"`
	Purpose string `json:"purpose"`
	Last4   string `json:"last4"`
}

func (p *CreateCardParams) Validate() error {
	eb := errs.B().Code(errs.InvalidArgument)
	if p.BankID == "" {
		return eb.Msg("bank_id is required").Err()
	}
	if p.Label == "" {
		return eb.Msg("label is required").Err()
	}
	if p.Purpose == "" {
		return eb.Msg("purpose is required").Err()
	}
	if len(p.Last4) != 4 {
		return eb.Msg("last4 must be exactly 4 digits").Err()
	}
	return nil
}

type ListCardsResponse struct {
	Cards []*Card `json:"cards"`
}

// GetCardByLast4AndBankParams is used by the private parser endpoint.
type GetCardByLast4AndBankParams struct {
	Last4    string `query:"last4"`
	BankSlug string `query:"bank_slug"`
}
