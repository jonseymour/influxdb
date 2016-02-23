package expvar

type Map struct {
}

type String struct {
}

type Var interface {
	String() string
}

type Int struct {
}

type Float struct {
}

type KeyValue struct {
	Key   string
	Value Var
}

func NewMap(n string) *Map {
	return nil
}

func NewInt(n string) *Int {
	return nil
}

func Get(n string) Var {
	return nil
}

func Do(func(kv KeyValue)) {

}

func (v *String) Set(s string) {
}

func (v *Int) Set(i int64) {
}

func (m *Map) Set(n string, v Var) {
}

func (m *Map) Init() {
}

func (m *Map) Add(n string, v int64) {}

func (m *Map) String() string {
	return ""
}

func (m *Map) Do(func(kv KeyValue)) {
}

func (m *Int) String() string {
	return ""
}

func (m *String) String() string {
	return ""
}

func (f *Float) String() string {
	return ""
}
