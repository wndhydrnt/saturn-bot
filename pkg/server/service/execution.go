package service

type Execution struct{}

func (w *Execution) EnqueueNew() {}

func (w *Execution) GetNext() {}

func (w *Execution) StoreResult() {}
