// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

// Command resize-measure reports the peak resident memory and wall time of a resize.
//
// It exists so that memory changes can be compared reproducibly: the same command run against an
// old and a new checkout produces numbers that are directly comparable. It is not part of CI.
//
//	go run ./cmd/resize-measure -src 6000x4000 -dst 1200x800 -type ycbcr420
//
// Peak RSS is read from /proc/self/status (VmHWM) and is therefore Linux-only; on other platforms
// the line is omitted and only the Go heap statistics are printed.
package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schankst/resize"
)

func main() {
	src := flag.String("src", "6000x4000", "source size, WxH")
	dst := flag.String("dst", "1200x800", "destination size, WxH")
	kind := flag.String("type", "ycbcr420", "source type: rgba, nrgba, gray, ycbcr420, ycbcr422, ycbcr444")
	interp := flag.String("kernel", "lanczos3", "interpolation: nearest, bilinear, bicubic, mitchell, lanczos2, lanczos3")
	repeat := flag.Int("repeat", 1, "sequential resizes per worker")
	parallel := flag.Int("parallel", 1, "concurrent workers")
	flag.Parse()

	srcW, srcH := mustSize(*src)
	dstW, dstH := mustSize(*dst)
	kernel, err := parseKernel(*interp)
	if err != nil {
		fail(err)
	}

	img, err := buildImage(*kind, srcW, srcH)
	if err != nil {
		fail(err)
	}

	// Let the source allocation settle before taking the baseline.
	runtime.GC()
	baseHeap, baseRSS := stats()

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < *parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < *repeat; j++ {
				out := resize.Resize(uint(dstW), uint(dstH), img, kernel)
				if out.Bounds().Empty() {
					panic("empty result")
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	heap, peakRSS := stats()

	fmt.Printf("source     %s %dx%d (%s)\n", *kind, srcW, srcH, humanBytes(sourceBytes(*kind, srcW, srcH)))
	fmt.Printf("dest       %dx%d %s, %d worker(s) x %d resize(s)\n", dstW, dstH, *interp, *parallel, *repeat)
	fmt.Printf("elapsed    %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("heap       %s -> %s (totalAlloc %s)\n", humanBytes(baseHeap), humanBytes(heap), humanBytes(totalAlloc()))
	if peakRSS > 0 {
		fmt.Printf("peak RSS   %s (baseline %s)\n", humanBytes(peakRSS), humanBytes(baseRSS))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "resize-measure:", err)
	os.Exit(2)
}

func mustSize(s string) (int, int) {
	parts := strings.Split(strings.ToLower(s), "x")
	if len(parts) != 2 {
		fail(fmt.Errorf("bad size %q, want WxH", s))
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil || w <= 0 {
		fail(fmt.Errorf("bad width in %q", s))
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil || h <= 0 {
		fail(fmt.Errorf("bad height in %q", s))
	}
	return w, h
}

func parseKernel(s string) (resize.InterpolationFunction, error) {
	switch strings.ToLower(s) {
	case "nearest":
		return resize.NearestNeighbor, nil
	case "bilinear":
		return resize.Bilinear, nil
	case "bicubic":
		return resize.Bicubic, nil
	case "mitchell":
		return resize.MitchellNetravali, nil
	case "lanczos2":
		return resize.Lanczos2, nil
	case "lanczos3":
		return resize.Lanczos3, nil
	}
	return 0, fmt.Errorf("unknown kernel %q", s)
}

func sourceBytes(kind string, w, h int) uint64 {
	switch kind {
	case "gray":
		return uint64(w * h)
	case "ycbcr420", "ycbcr422":
		return uint64(w*h*3/2 + 2)
	case "ycbcr444":
		return uint64(w * h * 3)
	default:
		return uint64(w * h * 4)
	}
}

func buildImage(kind string, w, h int) (image.Image, error) {
	// Deterministic, non-uniform content so the resize cannot be optimised away.
	switch strings.ToLower(kind) {
	case "rgba":
		m := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := y*m.Stride + x*4
				m.Pix[i+0] = uint8(x * 7)
				m.Pix[i+1] = uint8(y * 5)
				m.Pix[i+2] = uint8(x ^ y)
				m.Pix[i+3] = 0xff
			}
		}
		return m, nil
	case "nrgba":
		m := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := y*m.Stride + x*4
				m.Pix[i+0] = uint8(x * 7)
				m.Pix[i+1] = uint8(y * 5)
				m.Pix[i+2] = uint8(x ^ y)
				m.Pix[i+3] = uint8(x + y)
			}
		}
		return m, nil
	case "gray":
		m := image.NewGray(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.Pix[y*m.Stride+x] = uint8(x ^ y)
			}
		}
		return m, nil
	case "ycbcr420", "ycbcr422", "ycbcr444":
		var ratio image.YCbCrSubsampleRatio
		switch strings.ToLower(kind) {
		case "ycbcr420":
			ratio = image.YCbCrSubsampleRatio420
		case "ycbcr422":
			ratio = image.YCbCrSubsampleRatio422
		default:
			ratio = image.YCbCrSubsampleRatio444
		}
		m := image.NewYCbCr(image.Rect(0, 0, w, h), ratio)
		for i := range m.Y {
			m.Y[i] = uint8(i * 31)
		}
		for i := range m.Cb {
			m.Cb[i] = uint8(i * 17)
			m.Cr[i] = uint8(i * 13)
		}
		return m, nil
	}
	return nil, fmt.Errorf("unknown type %q", kind)
}

// stats returns the live Go heap and, on Linux, the peak resident set size of the process.
func stats() (heap uint64, peakRSS uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heap = m.HeapAlloc
	peakRSS = readRSS("VmHWM")
	return heap, peakRSS
}

func totalAlloc() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc
}

// readRSS reads a field such as VmHWM or VmRSS from /proc/self/status, in bytes. Returns 0 when the
// file or field is unavailable.
func readRSS(field string) uint64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, field+":") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}

func humanBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d B", v)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	f := float64(v)
	for _, u := range units {
		f /= unit
		if f < unit {
			return fmt.Sprintf("%.1f %s", f, u)
		}
	}
	return fmt.Sprintf("%.1f PiB", f/unit)
}
