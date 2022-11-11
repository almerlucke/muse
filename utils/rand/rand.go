package rand

import (
	"math"
	"math/rand"
)

const (
	defaultSeed uint64 = 1
)

type Rand struct {
	u, v, w uint64
}

func NewRand() *Rand {
	return NewRandWithSeed(defaultSeed)
}

func NewRandWithSeed(seed uint64) *Rand {
	r := &Rand{}
	r.seed(seed)
	return r
}

func (r *Rand) seed(s uint64) {
	r.v = 4101842887655102017
	r.w = 1
	r.u = s ^ r.v
	r.RandInt()
	r.v = r.u
	r.RandInt()
	r.w = r.v
	r.RandInt()
}

// RandInt Random number between [0, 2^64 - 1] (Numerical Recipes in C" Third Edition)
func (r *Rand) RandInt() uint64 {
	r.u = r.u*2862933555777941757 + 7046029254386353087
	r.v ^= r.v >> 17
	r.v ^= r.v << 31
	r.v ^= r.v >> 8
	r.w = 4294957665*(r.w&0xffffffff) + (r.w >> 32)
	x := r.u ^ (r.u << 21)
	x ^= x >> 35
	x ^= x << 4

	return (x + r.v) ^ r.w
}

// RandFloat Random number between [0, 1)
func (r *Rand) RandFloat() float64 {
	return 5.42101086242752217e-20 * float64(r.RandInt())
}

// distribution 0 -> center
// distribution 1 -> center +- width random 100%
func randCenter(center float64, width float64, distribution float64) float64 {
	mult := 0.0

	if distribution >= 0.001 {
		mult = 1.0 - math.Pow(rand.Float64(), distribution)
	}

	left := rand.Intn(2)
	out := 0.0
	if left == 0 {
		out = center - width*0.5*mult
	} else {
		out = center + width*0.5*mult
	}

	return out
}

type Cluster struct {
	Center float64
	Width  float64
}

func (c *Cluster) Rand(distribution float64) float64 {
	return randCenter(c.Center, c.Width, distribution)
}

type Clusters []*Cluster

func (clusters Clusters) Rand(clusterDistribution float64, valueDistribution float64) float64 {
	nf := float64(len(clusters))
	clusterIndex := int(randCenter(nf/2.0, nf, clusterDistribution))
	cluster := clusters[clusterIndex]
	return cluster.Rand(valueDistribution)
}

type ClusterRand struct {
	clusters            Clusters
	center              float64
	width               float64
	density             float64
	valueDistribution   float64
	clusterDistribution float64
	needUpdate          bool
}

func NewClusterRand(center float64, width float64, density float64, valueDistribution float64, clusterDistribution float64) *ClusterRand {
	cl := &ClusterRand{}
	cl.SetCenter(center)
	cl.SetWidth(width)
	cl.SetDensity(density)
	cl.SetValueDistribution(valueDistribution)
	cl.SetClusterDistribution(clusterDistribution)
	cl.Update()
	return cl
}

func (c *ClusterRand) Update() {
	if !c.needUpdate {
		return
	}

	n := len(c.clusters)

	clusterWidth := c.width / float64(n)
	centerClusterIndex := n / 2
	numSideClusters := (n - 1) / 2

	c.clusters[centerClusterIndex] = &Cluster{Center: c.center, Width: clusterWidth}

	for i := 0; i < numSideClusters; i++ {
		j := i + 1
		c.clusters[centerClusterIndex-j] = &Cluster{Center: c.center - clusterWidth*float64(j), Width: clusterWidth}
		c.clusters[centerClusterIndex+j] = &Cluster{Center: c.center + clusterWidth*float64(j), Width: clusterWidth}
	}

	c.needUpdate = false
}

func (c *ClusterRand) NeedUpdate() bool {
	return c.needUpdate
}

func (c *ClusterRand) setNumClusters(n int) {
	if c.clusters != nil && n == len(c.clusters) {
		return
	}

	c.clusters = make(Clusters, n)
	c.needUpdate = true
}

func (c *ClusterRand) Density() float64 {
	return c.density
}

func (c *ClusterRand) SetDensity(d float64) {
	c.density = d
	c.setNumClusters(1 + int(d*10.0)*2)
}

func (c *ClusterRand) Center() float64 {
	return c.center
}

func (c *ClusterRand) SetCenter(center float64) {
	c.center = center
	c.needUpdate = true
}

func (c *ClusterRand) Width() float64 {
	return c.width
}

func (c *ClusterRand) SetWidth(width float64) {
	c.width = width
	c.needUpdate = true
}

func (c *ClusterRand) ValueDistribution() float64 {
	return c.valueDistribution
}

func (c *ClusterRand) SetValueDistribution(dist float64) {
	c.valueDistribution = dist
}

func (c *ClusterRand) ClusterDistribution() float64 {
	return c.clusterDistribution
}

func (c *ClusterRand) SetClusterDistribution(dist float64) {
	c.clusterDistribution = dist
}

func (c *ClusterRand) Rand() float64 {
	return c.clusters.Rand(c.clusterDistribution, c.valueDistribution)
}
