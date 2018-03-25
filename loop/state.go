package loop

type State struct {

}

func AtomicGetState() (*State) {
	return &State{}
}

func (s *State) AtomicEnsureAttempt(svc *Service) (error) {
	
	return nil
}