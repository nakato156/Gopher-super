package engine

// PartialAccResult representa el resultado parcial del cálculo de similitudes
// sobre un subconjunto de usuarios. La idea es que estos resultados se
// agreguen/mergeen en un coordinador para luego reconstruir la matriz de
// similitud global.
type PartialAccResult struct {
	Dot  map[uint64]float64 `json:"dot"`
	Co   map[uint64]int     `json:"co"`
	Norm map[int]float64    `json:"norm"`
}

// ComputePartialSimilarities procesa solo un subconjunto de usuarios y devuelve
// las acumulaciones parciales (dot/co/norm) tal como las produce processUserBlock
// y mergeAcc, pero sin construir la SimMatrix completa. Este es el entry-point
// pensado para ejecutarse en un worker que recibe un bloque de usuarios.
func ComputePartialSimilarities(userSubset []int, userRatings map[int]map[int]float64) PartialAccResult {
	// params:
	//   - userSubset: lista de IDs de usuario que este nodo debe procesar
	//   - userRatings: mapa global de ratings (userID -> movieID -> rating)
	//
	// return:
	//   - PartialAccResult con los mapas parciales listos para ser mergeados en el coordinador.

	local := newPartialAcc()
	buf := make([]int, 0, 64)
	processUserBlock(local, userSubset, userRatings, buf)

	return PartialAccResult{
		Dot:  local.dot,
		Co:   local.co,
		Norm: local.norm,
	}
}
