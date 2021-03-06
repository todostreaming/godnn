package main

import (
	// "github.com/xigh/godnn"
	"../../../godnn"
	"fmt"
	"log"
	"os"
)

const (
	rate = 0.15
	minErr = 0.001
	maxIter = 100000
)

var (
	topology = []uint{ 6, 30, 12, 2 }
)

const (
	Rock = iota
	Paper
	Scissors
	N
)

var (
	names = []string{ "Rock", "Paper", "Scissors" }

	COLOR_RESET = "\x1b[39;49;0m"
	COLOR_RED = "\x1b[31;1m"
	COLOR_GREEN = "\x1b[32;1m"
	COLOR_YELLOW = "\x1b[33;1m"
	COLOR_OTHER = "\x1b[34;1m"
	COLOR_DBLUE = "\x1b[35;1m"
	COLOR_BLUE = "\x1b[36;1m"
	COLOR_WHITE = "\x1b[37;1m"
)

func eval(a, b int) []float64 {
	// paper covers rock
	if a == Paper && b == Rock {
		return []float64{ 1, 0 }
	}
	if b == Paper && a == Rock {
		return []float64{ 0, 1 }
	}

	// rock cruches scissor
	if a == Rock && b == Scissors {
		return []float64{ 1, 0 }
	}
	if b == Rock && a == Scissors {
		return []float64{ 0, 1 }
	}

	// scissors cuts paper
	if a == Scissors && b == Paper {
		return []float64{ 1, 0 }
	}
	if b == Scissors && a == Paper {
		return []float64{ 0, 1 }
	}

	return []float64{ 0, 0 }
}

func test(net *dnn.Net) {
	avg := 0.0
	nb := 0
	for a := 0; a < N; a += 1 {
		for b := 0; b < N; b += 1 {
			input := make([]float64, N * 2)
			input[a] = 1
			input[b + N] = 1

			results, err := net.Predict(input)
			if err != nil {
				log.Fatal(err)
			}

			output := eval(a, b)
			
			if len(output) != len(results) {
				log.Fatal("ouput size != results size")
			}

			dist := 0.0
			for n := range output {
				delta := output[n] - results[n]
				dist += delta * delta
			}
			dist /= float64(len(output))
			avg += dist
			nb += 1

			fmt.Printf(COLOR_BLUE + "%-10s" + COLOR_RESET, names[a])
			fmt.Printf(" vs ")
			fmt.Printf(COLOR_BLUE + "%-10s" + COLOR_RESET, names[b])
			fmt.Printf(" src=" + COLOR_YELLOW + "%v" + COLOR_RESET, input)
			fmt.Printf(" res=" + COLOR_OTHER + "%12.7f" + COLOR_RESET, results)
			fmt.Printf(" exp=" + COLOR_GREEN + "%v" + COLOR_RESET, output)
			fmt.Printf(" err=" + COLOR_RED + "%12.7f%%\n" + COLOR_RESET, dist * 100)
		}
	}
	fmt.Printf("average error: %9.5f%%\n\n", 100 * avg / float64(nb))
}

func train(net *dnn.Net, min float64, max uint64) uint64 {
	i := uint64(0)
	p := 0.0
	for {
		if i % 1000 == 0 {
			fmt.Printf(".")
		}
		i += 1

		if i > 0 && i % 5000 == 0 {
			fmt.Printf("%.2f", p)
		}

		if i > max {
			fmt.Printf("\ntoo many iterations\n")
			break
		}
		
		avg := 0.0
		nb := 0
		for a := 0; a < N; a += 1 {
			for b := 0; b < N; b += 1 {
				input := make([]float64, N * 2)
				input[a] = 1
				input[b + N] = 1
				
				output := eval(a, b)
				
				dist, err := net.Train(input, output, rate)
				if err != nil {
					log.Fatal(err)
				}
				
				avg += dist
				nb += 1
			}
		}

		p = 100 * avg / float64(nb)
		if (p < min) {
			fmt.Printf("\naverage error=%9.5f%%\n", avg)
			break
		}
	}
	return i
}

func learnAndSave(name string) error {
	net, err := dnn.Create(topology)
	if err != nil {
		return err
	}

	// -------
	
	fmt.Printf("topology: %v\n", topology)

	// -------
	
	fmt.Printf("test before training:\n")
	test(net)

	// -------
	
	fmt.Printf("learning [min avg error: %f]:\n", minErr)
	itn := train(net, minErr, maxIter)
	fmt.Printf(" - %d iterations\n\n", itn)

	// -------

	fmt.Printf("test after training:\n")
	test(net);

	w, err := os.Create(name)
	if err != nil {
		return err
	}
	defer w.Close()
	
	return net.Save(w)
}

func loadAndTest(name string) error {
	r, err := os.Open(name)
	if err != nil {
		return err
	}
	defer r.Close()
	
	net, err := dnn.Load(r)
	if err != nil {
		return err
	}

	// -------
	
	fmt.Printf("topology: %v\n", net.Topology())

	// -------
	
	fmt.Printf("test before training:\n")
	test(net)

	return nil
}

func main() {
	err := learnAndSave("network.json")
	if err != nil {
		log.Fatal(err)
	}

	err = loadAndTest("network.json")
	if err != nil {
		log.Fatal(err)
	}
}
