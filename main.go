func main(){
	k int = runtime.NumCPU()

	A := make([]float64, vectorSize)
	// fill Vector A
	for i := 0; i < vectorSize; i++ {
		A[i] = rand.Float64()
	}
	
}