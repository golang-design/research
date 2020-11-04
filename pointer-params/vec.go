package pparam

type vec struct {
	x, y, z, w float64
}

func (v vec) addv(u vec) vec {
	return vec{v.x + u.x, v.y + u.y, v.z + u.z, v.w + u.w}
}

func (v *vec) addp(u *vec) *vec {
	v.x, v.y, v.z, v.w = v.x+u.x, v.y+u.y, v.z+u.z, v.w+u.w
	return v
}
