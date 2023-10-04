package main

import (
	"math"
	"math/rand"
	"sort"

	"github.com/almerlucke/muse/plot"
	"gonum.org/v1/plot/plotter"
)

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

func main() {
	cl := NewClusterRand(3.0, 2.0, 0.4, 0.3, 0.8)

	accuracy := 10000.0
	n := 2000000
	buckets := map[int]int{}

	for i := 0; i < n; i++ {
		v := cl.Rand()
		buckets[int(v*accuracy)]++
	}

	points := make(plotter.XYs, len(buckets))
	index := 0
	for k, v := range buckets {
		points[index].X = float64(k) / accuracy
		points[index].Y = float64(v)
		index++
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].X < points[j].X
	})

	// sort.Sort(sort.Float64Slice(v))

	plot.PlotPoints(points, 500, 300, "/Users/almerlucke/Desktop/exponential.png")
}
