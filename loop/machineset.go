package loop

type MemoryMachineSet struct {

}

func NewMemoryMachineSetFromCRD() (*MemoryMachineSet, error) {
	return &MemoryMachineSet{}, nil
}

func (mm *MemoryMachineSet) Ensure() (error) {
	return nil
}