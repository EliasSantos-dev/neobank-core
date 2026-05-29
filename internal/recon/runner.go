package recon

import "context"

// Source fornece os dados agregados para a reconciliação (implementado por *store.Store).
type Source interface {
	ReconInput(ctx context.Context) (Input, error)
}

// Run coleta os dados e aplica Check.
func Run(ctx context.Context, src Source) (Report, error) {
	in, err := src.ReconInput(ctx)
	if err != nil {
		return Report{}, err
	}
	return Check(in), nil
}
