// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"reflect"
	"sync"

	"azul3d.org/engine/lmath"
)

// Boundable represents any object that can return it's axis-aligned bounding
// box.
type Boundable interface {
	// Bounds returns the axis-aligned bounding box of this boundable object.
	Bounds() lmath.Rect3
}

// Bounds is a simple datatype which implements the Boundable interface.
type Bounds lmath.Rect3

// Bounds implements the Boundable interface.
func (b Bounds) Bounds() lmath.Rect3 {
	return lmath.Rect3(b)
}

// TexCoordSet represents a single texture coordinate set for a mesh.
type TexCoordSet struct {
	// The slice of texture coordinates for the set.
	Slice []TexCoord

	// Whether or not the texture coordinates of this set have changed since
	// the last time the mesh was loaded. If set to true the device should
	// take note and re-upload the data slice to the graphics hardware.
	Changed bool
}

// VertexAttrib represents a per-vertex attribute.
type VertexAttrib struct {
	// The literal per-vertex data slice. It must be a slice whose length is
	// exactly the same as the mesh's Vertices slice (because it is literally
	// per-vertex data). The underlying type must be one of the following or
	// else the attribute may be ignored completely:
	//  []float32
	//  [][]float32
	//  []gfx.Vec3
	//  [][]gfx.Vec3
	//  []gfx.Vec4
	//  [][]gfx.Vec4
	//  []gfx.Mat4
	//  [][]gfx.Mat4
	//  []gfx.Color
	//  [][]gfx.Color
	//  []gfx.TexCoord
	//  [][]gfx.TexCoord
	Data interface{}

	// Whether or not the per-vertex data (see the Data field) has changed
	// since the last time the mesh was loaded. If set to true the device
	// should take note and re-upload the data slice to the device's memory.
	Changed bool
}

// Copy returns a new copy of this vertex attribute data set. It makes a deep
// copy of the underlying Data slice. Explicitly not copied is the Changed
// boolean.
func (a VertexAttrib) Copy() VertexAttrib {
	var cpy interface{}
	switch t := a.Data.(type) {
	case []float32:
		c := make([]float32, len(t))
		copy(c, t)
		cpy = c

	case []TexCoord:
		c := make([]TexCoord, len(t))
		copy(c, t)
		cpy = c

	case []Vec3:
		c := make([]Vec3, len(t))
		copy(c, t)
		cpy = c

	case []Vec4:
		c := make([]Vec4, len(t))
		copy(c, t)
		cpy = c

	case []Color:
		c := make([]Color, len(t))
		copy(c, t)
		cpy = c

	case []Mat4:
		c := make([]Mat4, len(t))
		copy(c, t)
		cpy = c

	case [][]float32:
		c := make([][]float32, len(t))
		for i, s := range t {
			c[i] = make([]float32, len(s))
			copy(c[i], t[i])
		}

	case [][]TexCoord:
		c := make([][]TexCoord, len(t))
		for i, s := range t {
			c[i] = make([]TexCoord, len(s))
			copy(c[i], t[i])
		}

	case [][]Vec3:
		c := make([][]Vec3, len(t))
		for i, s := range t {
			c[i] = make([]Vec3, len(s))
			copy(c[i], t[i])
		}

	case [][]Vec4:
		c := make([][]Vec4, len(t))
		for i, s := range t {
			c[i] = make([]Vec4, len(s))
			copy(c[i], t[i])
		}

	case [][]Color:
		c := make([][]Color, len(t))
		for i, s := range t {
			c[i] = make([]Color, len(s))
			copy(c[i], t[i])
		}

	case [][]Mat4:
		c := make([][]Mat4, len(t))
		for i, s := range t {
			c[i] = make([]Mat4, len(s))
			copy(c[i], t[i])
		}

	default:
		return VertexAttrib{}
	}
	return VertexAttrib{Data: cpy}
}

// Mesh represents a single mesh made up of several components. A mesh may or
// may not be made up of indexed vertices, etc, depending on whether or not
// len(m.Indices) == 0 holds true.
//
// In the event that a mesh is indexed, m.Indices holds the indices and it can
// be expected that each other slice (Vertices for instance) will hold at least
// enough elements (or be nil) such that the each index will not be out of
// bounds.
//
// A mesh and it's methods are not safe for access from multiple goroutines
// concurrently.
type Mesh struct {
	// The native object of this mesh. Once the mesh is loaded by a device this
	// field will be initialized by the device. Only device implementations
	// should assign values to this field.
	NativeMesh Destroyable

	// Whether or not this mesh is currently loaded or not.
	Loaded bool

	// The primitive type of this mesh. The zero-value (i.e. default value) is
	// Triangles.
	Primitive

	// If true then when this mesh is loaded the sources of it will be kept
	// instead of being set to nil (which allows them to be garbage collected).
	KeepDataOnLoad bool

	// Dynamic is a hint (it does not restrict how the mesh may be used) to the
	// graphics device on how this mesh might be used. If you intend to update
	// mesh data often (i.e. it's not static) then set this to true.
	Dynamic bool

	// AABB is the axis aligned bounding box of this mesh. There may not be one
	// if AABB.Empty() == true, but one can be calculate using the
	// CalculateBounds() method.
	AABB lmath.Rect3

	// A slice of indices, if non-nil then this slice contains indices into
	// each other slice (such as Vertices) and this is a indexed mesh.
	//
	// The indices are uint32 (instead of int) for compatability with graphics
	// hardware.
	Indices []uint32

	// Whether or not the indices have changed since the last time the mesh
	// was loaded. If set to true the device should take note and re-upload the
	// data slice to the graphics hardware.
	IndicesChanged bool

	// The slice of vertices for the mesh.
	Vertices []Vec3

	// Whether or not the vertices have changed since the last time the
	// mesh was loaded. If set to true the device should take note and
	// re-upload the data slice to the graphics hardware.
	VerticesChanged bool

	// The slice of vertex colors for the mesh.
	Colors []Color

	// Whether or not the vertex colors have changed since the last time
	// the mesh was loaded. If set to true the device should take note
	// and re-upload the data slice to the graphics hardware.
	ColorsChanged bool

	// The slice of normals for the mesh.
	Normals []Vec3

	// Whether or not the normals have changed since the last time the
	// mesh was loaded. If set to true the device should take note and
	// re-upload the data slice to the graphics hardware.
	NormalsChanged bool

	// A slice of barycentric coordinates for the mesh.
	Bary []Vec3

	// Whether or not the barycentric coordinates have changed since the last
	// time the mesh was loaded. If set to true the device should take note
	// and re-upload the data slice to the graphics hardware.
	BaryChanged bool

	// A slice of texture coordinate sets for the mesh, there may be
	// multiple sets which directly relate to multiple textures on a
	// object.
	TexCoords []TexCoordSet

	// A map of custom per-vertex attributes for the mesh. It is analogous to
	// the Colors, Bary, and TexCoords fields. It allows you to submit a set of
	// named custom per-vertex data to shaders.
	//
	// For instance you could submit a set of per-vertex vec3's with:
	//  myData := make([]gfx.Vec3, len(mesh.Vertices))
	//  mesh.Attribs["MyName"] = gfx.VertexAttrib{
	//      Data: myData,
	//  }
	//
	// If changes to the data are made, the data set will have to be uploaded
	// to the graphics hardware again, so you must inform the device when you
	// change the data:
	//  ... modify myData ...
	//  mesh.Attribs["MyName"].Changed = true
	//
	// In GLSL you could access that per-vertex data by writing:
	//  attribute vec3 MyName;
	//
	// Arrays of data are available in GLSL by slice indice suffixes:
	//  // Data declared in Go:
	//  myData := make([][]gfx.Mat4, 2)
	//
	//  // And in GLSL:
	//  attribute mat4 MyName0; // Per-vertex data from myData[0].
	//  attribute mat4 MyName1; // Per-vertex data from myData[1].
	//
	// See the documentation on the VertexAttrib type for more information
	// regarding what data types may be used.
	Attribs map[string]VertexAttrib
}

// Copy returns a new copy of this Mesh. Depending on how large the mesh is
// this may be an expensive operation. Explicitly not copied over is the native
// mesh, the OnLoad slice, and the loaded and changed statuses (Loaded,
// IndicesChanged, VerticesChanged, etc).
func (m *Mesh) Copy() *Mesh {
	cpy := &Mesh{
		nil,   // Native mesh -- not copied.
		false, // Loaded status -- not copied.
		m.Primitive,
		m.KeepDataOnLoad,
		m.Dynamic,
		m.AABB,
		make([]uint32, len(m.Indices)),
		false, // IndicesChanged -- not copied.
		make([]Vec3, len(m.Vertices)),
		false, // VerticesChanged -- not copied.
		make([]Color, len(m.Colors)),
		false, // ColorsChanged -- not copied.
		make([]Vec3, len(m.Normals)),
		false, // NormalsChanged -- not copied.
		make([]Vec3, len(m.Bary)),
		false, // BaryChanged -- not copied.
		make([]TexCoordSet, len(m.TexCoords)),
		make(map[string]VertexAttrib, len(m.Attribs)),
	}

	copy(cpy.Indices, m.Indices)
	copy(cpy.Vertices, m.Vertices)
	copy(cpy.Colors, m.Colors)
	copy(cpy.Normals, m.Normals)
	copy(cpy.Bary, m.Bary)
	for index, set := range m.TexCoords {
		setCpy := TexCoordSet{
			Slice: make([]TexCoord, len(set.Slice)),
		}
		copy(setCpy.Slice, set.Slice)
		cpy.TexCoords[index] = setCpy
	}
	for name, attrib := range m.Attribs {
		cpy.Attribs[name] = attrib.Copy()
	}
	return cpy
}

// Bounds implements the Boundable interface. If the AABB of this mesh is empty
// then the bounds are calculated.
func (m *Mesh) Bounds() lmath.Rect3 {
	if m.AABB.Empty() {
		m.CalculateBounds()
	}
	bounds := m.AABB
	return bounds
}

// GenerateBary generates the barycentric coordinates for this mesh.
func (m *Mesh) GenerateBary() {
	var (
		bci = -1
		v   Vec3
	)
	for _ = range m.Vertices {
		// Add barycentric coordinates.
		bci++
		switch bci % 3 {
		case 0:
			v = Vec3{1, 0, 0}
		case 1:
			v = Vec3{0, 1, 0}
		case 2:
			v = Vec3{0, 0, 1}
		}
		m.Bary = append(m.Bary, v)
	}
}

// CalculateBounds calculates a new axis aligned bounding box for this mesh.
func (m *Mesh) CalculateBounds() {
	var bb lmath.Rect3
	if len(m.Vertices) > 0 {
		for _, v32 := range m.Vertices {
			v := v32.Vec3()
			bb.Min = bb.Min.Min(v)
			bb.Max = bb.Max.Max(v)
		}
	}
	m.AABB = bb
}

// HasChanged tells if any of the data slices of the mesh are marked as having
// changed.
func (m *Mesh) HasChanged() bool {
	if m.IndicesChanged || m.VerticesChanged || m.ColorsChanged || m.NormalsChanged || m.BaryChanged {
		return true
	}
	for _, texCoordSet := range m.TexCoords {
		if texCoordSet.Changed {
			return true
		}
	}
	for _, attrib := range m.Attribs {
		if attrib.Changed {
			return true
		}
	}
	return false
}

// fastAppendSlice is basically reflect.AppendSlice, but it implements fast
// paths for common types.
func fastAppendSlice(a, b interface{}) interface{} {
	switch aa := a.(type) {
	case []float32:
		return append(aa, b.([]float32)...)
	case [][]float32:
		return append(aa, b.([][]float32)...)
	case []Vec3:
		return append(aa, b.([]Vec3)...)
	case [][]Vec3:
		return append(aa, b.([][]Vec3)...)
	case []Mat4:
		return append(aa, b.([]Mat4)...)
	case [][]Mat4:
		return append(aa, b.([][]Mat4)...)
	case []Vec4:
		return append(aa, b.([]Vec4)...)
	case [][]Vec4:
		return append(aa, b.([][]Vec4)...)
	case []Color:
		return append(aa, b.([]Color)...)
	case [][]Color:
		return append(aa, b.([][]Color)...)
	default:
		return reflect.AppendSlice(reflect.ValueOf(a), reflect.ValueOf(b)).Interface()
	}
}

// Append appends the other mesh's data slices to this one.
//
// This function can properly maintain the existing index type of a mesh, so
// that in any of the following cases:
//
//  mesh.Append(mesh)
//  indexedMesh.Append(indexedMesh)
//  indexedMesh.Append(mesh)
//  mesh.Append(indexedMesh)
//
// The original mesh's indexing is kept: anything appended to an indexed mesh
// always ends up as an indexed mesh, and vice versa.
//
// Only data slices that both meshes have in common are kept: e.g. if only one
// mesh has vertex colors, the mesh (m) will have it's vertex colors set to
// nil. You can check which data slices are in common by comparing the states
// of both meshes (see the MeshState documentation).
func (m *Mesh) Append(other *Mesh) {
	if len(other.Vertices) == 0 {
		// No vertices: nothing to do.
		return
	}

	// appendData appends a single data slice, y, to the given data slice, x.
	// It operates similar to:
	//
	//  m.Vertices = append(m.Vertices, other.Vertices...)
	//  ->
	//  val = appendData(m.Vertices, other.Vertices)
	//
	// The main difference is that fixIndices below sometimes modifies it's
	// behavior in order to generate correct indices for the data. By utilizing
	// appendData instead of append directly; indices are automatically fixed
	// below.
	var (
		appendData    = fastAppendSlice
		dataLenBefore = uint32(len(m.Vertices))
		fixIndices    = func() {}
	)
	if len(m.Indices) > 0 && len(other.Indices) > 0 {
		// i.e. append(IndexedMesh, IndexedMesh); For this case in order to fix
		// the indices we simply append each index of the second mesh offset by
		// the data length of the first mesh (i.e. len(m.Vertices)).
		fixIndices = func() {
			for _, index := range other.Indices {
				m.Indices = append(m.Indices, index+dataLenBefore)
			}
			m.IndicesChanged = true
		}

	} else if len(m.Indices) > 0 && len(other.Indices) == 0 {
		// i.e. append(IndexedMesh, Mesh); For this case we need to fill in
		// the missing indices (which in our case are simply offset counter
		// values).
		fixIndices = func() {
			for i := 0; i < len(other.Vertices); i++ {
				m.Indices = append(m.Indices, dataLenBefore+uint32(i))
			}
			m.IndicesChanged = true
		}

	} else if len(m.Indices) == 0 && len(other.Indices) > 0 {
		// i.e. append(Mesh, IndexedMesh); For this case we don't need to fix
		// the indices but we do need a different behavior from appendData;
		// instead of simply appending the data slices together it will expand
		// the slice data by looking at each index.
		appendData = func(xx, yy interface{}) interface{} {
			x := reflect.ValueOf(xx)
			y := reflect.ValueOf(yy)
			if y.Len() == 0 {
				// Be careful not to index a slice that does not have any data.
				return x.Interface()
			}
			for _, index := range other.Indices {
				x = reflect.Append(x, y.Index(int(index)))
			}
			return x.Interface()
		}
	}

	// Append vertices.
	m.Vertices = appendData(m.Vertices, other.Vertices).([]Vec3)
	m.VerticesChanged = true

	// Handle handles the appension of two data slices, one from each mesh. It
	// checks if the two data slices are differing in non-zero lengths (i.e. if
	// one has colors but the other does not).
	handle := func(x0, y0 interface{}, changed *bool) {
		x := reflect.ValueOf(x0).Elem()
		y := reflect.ValueOf(y0).Elem()
		xSlice := x
		ySlice := y
		if x.Kind() == reflect.Interface {
			xSlice = x.Elem()
		}
		if y.Kind() == reflect.Interface {
			ySlice = y.Elem()
		}
		if (xSlice.Len() > 0) != (ySlice.Len() > 0) {
			// The slices differ in non-zero lengths, e.g. one has colors but
			// the other does not.
			x.Set(reflect.Zero(x.Type()))
			*changed = true
			return
		}

		// Both slices are present, so append them together if needed.
		if ySlice.Len() > 0 {
			ap := appendData(xSlice.Interface(), ySlice.Interface())
			x.Set(reflect.ValueOf(ap))
			*changed = true
		}
	}

	// Handle each data slice.
	handle(&m.Colors, &other.Colors, &m.ColorsChanged)
	handle(&m.Normals, &other.Normals, &m.NormalsChanged)
	handle(&m.Bary, &other.Bary, &m.BaryChanged)

	// Handle texture coordinates.
	for i, tcs := range m.TexCoords {
		handle(&tcs.Slice, &other.TexCoords[i].Slice, &tcs.Changed)
		m.TexCoords[i] = tcs
	}

	// Handle vertex attributes.
	for name, attrib := range m.Attribs {
		otherAttrib := other.Attribs[name]
		handle(
			&attrib.Data,
			&otherAttrib.Data,
			&attrib.Changed,
		)
		m.Attribs[name] = attrib
	}

	// Now that we are done appending data, we fix the indices.
	fixIndices()
}

// ClearData sets the data slices of this mesh to nil if m.KeepDataOnLoad is
// set to false.
func (m *Mesh) ClearData() {
	if !m.KeepDataOnLoad {
		m.Indices = nil
		m.Vertices = nil
		m.Colors = nil
		m.Normals = nil
		m.Bary = nil
		m.TexCoords = nil
		m.Attribs = nil
	}
}

// Reset resets this mesh to it's default (NewMesh) state.
func (m *Mesh) Reset() {
	m.NativeMesh = nil
	m.Loaded = false
	m.Primitive = Triangles
	m.KeepDataOnLoad = false
	m.Dynamic = false
	m.AABB = lmath.Rect3Zero
	m.Indices = m.Indices[:0]
	m.IndicesChanged = false
	m.Vertices = m.Vertices[:0]
	m.VerticesChanged = false
	m.Colors = m.Colors[:0]
	m.ColorsChanged = false
	m.Normals = m.Normals[:0]
	m.NormalsChanged = false
	m.Bary = m.Bary[:0]
	m.BaryChanged = false
	for _, tcs := range m.TexCoords {
		tcs.Slice = nil
		tcs.Changed = false
	}
	m.TexCoords = m.TexCoords[:0]
	m.Attribs = make(map[string]VertexAttrib)
}

// Destroy destroys this mesh for use by other callees to NewMesh. You must not
// use it after calling this method. This makes an implicit call to
// m.NativeMesh.Destroy.
func (m *Mesh) Destroy() {
	if m.NativeMesh != nil {
		m.NativeMesh.Destroy()
	}
	m.Reset()
	meshPool.Put(m)
}

var meshPool = sync.Pool{
	New: func() interface{} {
		return &Mesh{
			Attribs: make(map[string]VertexAttrib),
		}
	},
}

// NewMesh returns a new *Mesh, for effeciency it may be a re-used one (see the
// Destroy method) whose slices have zero-lengths.
func NewMesh() *Mesh {
	return meshPool.Get().(*Mesh)
}
