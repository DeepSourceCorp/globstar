
import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
	
	// Insecure: Automatically exposes pprof endpoints over HTTP, allowing potential attackers to access profiling data.
	// <expect-error> automatic exposure of pprof endpoint
	_ "net/http/pprof"
)

func main() {
	// Insecure Example: 
	// Importing "net/http/pprof" without proper authentication exposes profiling endpoints (e.g., /debug/pprof/)
	// to anyone with network access. This can leak sensitive information like memory usage and stack traces.
	fmt.Println("Insecure pprof endpoint automatically exposed (not recommended).")

	// Secure Example:
	// Use "runtime/pprof" to manually control profiling without exposing HTTP endpoints.
	
	// Create a CPU profile file
	cpuProfile, err := os.Create("cpu_profile.prof")
	if err != nil {
		log.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuProfile.Close()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		log.Fatalf("Could not start CPU profile: %v", err)
	}
	fmt.Println("CPU profiling started...")

	// Simulate workload for profiling
	doWork()

	// Stop CPU profiling
	pprof.StopCPUProfile()
	fmt.Println("CPU profiling stopped. Profile saved to cpu_profile.prof")
}

func doWork() {
	// Simulated CPU-intensive task
	start := time.Now()
	for i := 0; i < 5000000; i++ {
		_ = i * i
	}
	fmt.Printf("Work completed in %v\n", time.Since(start))
}
