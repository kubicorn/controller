package loop


type MemoryMachine struct {

}

func NewMemoryMachineFromCRD() (*MemoryMachine, error) {
	return &MemoryMachine{}, nil
}

func (m *MemoryMachine) Ensure() (error) {
	return nil
}

