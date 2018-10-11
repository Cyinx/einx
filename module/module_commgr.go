package module

type ComponentMgr interface {
	OnComponentCreate(Context, ComponentID)
	OnComponentError(Context, error)
}
