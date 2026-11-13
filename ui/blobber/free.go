package blobber

import "math"

// No Grid (opt.NoGrid): the camera walks freely and looks around; the
// host's game still moves square by square. Walk keeps the camera inside
// the host's square, lets it lean out through an open side, and asks for
// the step once it crosses; the host's answer moves the square on (or,
// refused, the side closes until the camera backs off it).

// free is the free camera: its floor position, heading and pitch.
type free struct {
	on         bool
	x, z       float64
	yaw, pitch float64 // pitch: the horizon's shift, a share of the view's height
	cell       [2]int  // the host's square
	tried      int     // the side whose step the host was last asked for (-1: none)
}

// Free walking: top speed in squares a second, turn speed in radians a
// second at full input, the camera's radius against a wall, how far it
// leans out through an open side, and the pitch's reach.
const freeSpeed, freeTurn, freeRadius, freeLean, freeMaxPitch = 3.0, 2.6, 0.2, 0.45, 0.3

// FreeInput is one tick of free walking: forward and right (-1..1, the
// keys or a stick), a turn and a pitch (radians, the mouse's drag or the
// right stick).
type FreeInput struct{ Forward, Right, Turn, Pitch float64 }

// Walk advances the free camera by one tick (1/60 s) in the host's square
// at, whose sides open (N E S W) can be stepped through (open, or a door
// to try); ready: the host takes a move now, else the camera stands. It
// returns the square and facing the camera asks for: at itself, nothing.
func (v *View) Walk(at Camera, open [4]bool, in FreeInput, ready bool) Camera {
	f := &v.g.free
	sq := [2]int{at.X, at.Y}
	cx, cz := float64(at.X)+0.5, float64(at.Y)+0.5
	switch d := [2]int{sq[0] - f.cell[0], sq[1] - f.cell[1]}; {
	case !f.on || d[0]*d[0]+d[1]*d[1] > 1: // a jump (an area, a teleport, the first tick): to the square's middle
		*f = free{on: true, x: cx, z: cz, yaw: float64(at.Dir) * math.Pi / 2, cell: sq, tried: -1}
	case d != [2]int{}: // the step came: the camera steps into the square if it was not yet
		f.cell, f.tried = sq, -1
		f.x = math.Max(float64(sq[0])+0.01, math.Min(float64(sq[0])+0.99, f.x))
		f.z = math.Max(float64(sq[1])+0.01, math.Min(float64(sq[1])+0.99, f.z))
	}
	if !ready {
		return v.asks()
	}
	const dt = 1.0 / 60
	f.yaw = math.Mod(f.yaw+in.Turn, 2*math.Pi)
	f.pitch = math.Max(-freeMaxPitch, math.Min(freeMaxPitch, f.pitch+in.Pitch))
	fw, rt := in.Forward, in.Right
	if n := math.Hypot(fw, rt); n > 1 {
		fw, rt = fw/n, rt/n
	}
	s := freeSpeed * math.Max(MinSpeed, math.Min(MaxSpeed, opt.Speed)) * dt
	f.x += (math.Sin(f.yaw)*fw + math.Cos(f.yaw)*rt) * s
	f.z += (-math.Cos(f.yaw)*fw + math.Sin(f.yaw)*rt) * s
	// the square's sides: a closed one (or the one just refused) keeps the camera its radius off it
	lo, hi := [2]float64{float64(sq[0]), float64(sq[1])}, [2]float64{float64(sq[0] + 1), float64(sq[1] + 1)}
	p := [2]*float64{&f.x, &f.z}
	for side, d := range dirs {
		a := 0 // the axis the side is across
		if d[0] == 0 {
			a = 1
		}
		k := d[a]
		edge := lo[a]
		if k > 0 {
			edge = hi[a]
		}
		reach := freeLean
		if !open[side] || f.tried == side {
			reach = -freeRadius
		}
		if over := (*p[a] - edge) * float64(k); over > reach {
			*p[a] = edge + float64(k)*reach
		}
		if f.tried == side && (edge-*p[a])*float64(k) > 0.3 { // backed off the refused side: it may be tried again
			f.tried = -1
		}
	}
	// out through two sides at once (a corner): only the one it is further through
	ox, oz := math.Max(lo[0]-f.x, f.x-hi[0]), math.Max(lo[1]-f.z, f.z-hi[1])
	if ox > 0 && oz > 0 {
		if ox > oz {
			f.z = math.Max(lo[1], math.Min(hi[1]-1e-9, f.z))
		} else {
			f.x = math.Max(lo[0], math.Min(hi[0]-1e-9, f.x))
		}
	}
	c := v.asks()
	if c.X != at.X || c.Y != at.Y {
		for side, d := range dirs {
			if d == [2]int{c.X - at.X, c.Y - at.Y} {
				f.tried = side
			}
		}
	}
	return c
}

// asks is the square the free camera stands on and the facing nearest its heading.
func (v *View) asks() Camera {
	f := &v.g.free
	dir := int(math.Round(f.yaw/(math.Pi/2))) % 4
	if dir < 0 {
		dir += 4
	}
	return Camera{int(math.Floor(f.x)), int(math.Floor(f.z)), dir}
}

// FreeOff ends free walking: the next Walk starts it again in the middle of the host's square.
func (v *View) FreeOff() { v.g.free.on = false }
