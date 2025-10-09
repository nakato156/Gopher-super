package similarity

import (
	"context"
	"sync"
)

type Service struct {
	M        *Matrix
	In       chan Request
	OutBatch chan []Request
	Workers  int
	Block    int
}

// -------- Métodos de Service --------
func (s *Service) Submit(ctx context.Context, A []uint32) ([]float64, bool) {
	cpy := make([]uint32, len(A))
	copy(cpy, A)
	rep := make(chan []float64, 1)
	req := Request{A: cpy, Reply: rep, Ctx: ctx}

	select {
	case s.In <- req:
	case <-ctx.Done():
		return nil, false
	}
	select {
	case res := <-rep:
		return res, true
	case <-ctx.Done():
		return nil, false
	}
}

func (s *Service) LoopBatches(process func([]Request)) {
	for batch := range s.OutBatch {
		process(batch)
	}
}

func (s *Service) ProcessBatch(batch []Request, compute func(req Request, row Row) float64) {
	if len(batch) == 0 {
		return
	}
	rows := s.M.Rows
	outs := make([][]float64, len(batch))
	for i := range batch {
		outs[i] = make([]float64, len(rows))
	}

	type task struct{ lo, hi int }
	jobs := make(chan task, s.Workers*32)

	var wg sync.WaitGroup
	wg.Add(s.Workers)
	for w := 0; w < s.Workers; w++ {
		go func() {
			defer wg.Done()
			for t := range jobs {
				for r := t.lo; r < t.hi; r++ {
					row := rows[r]
					for i := range batch {
						select {
						case <-batch[i].Ctx.Done():
							continue
						default:
						}
						outs[i][r] = compute(batch[i], row)
					}
				}
			}
		}()
	}

	blk := s.Block
	if blk < 16 {
		blk = 16
	}
	if blk > len(rows) {
		blk = len(rows)
	}
	for start := 0; start < len(rows); start += blk {
		end := start + blk
		if end > len(rows) {
			end = len(rows)
		}
		jobs <- task{lo: start, hi: end}
	}
	close(jobs)
	wg.Wait()

	for i, req := range batch {
		select {
		case <-req.Ctx.Done():
		default:
			req.Reply <- outs[i]
		}
		close(req.Reply)
	}
}
