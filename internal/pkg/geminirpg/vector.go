package geminirpg

import (
	"gonum.org/v1/gonum/mat"
)

func CosineSimilarity(a, b []float32) float64 {
	vecA := mat.NewVecDense(len(a), float32To64(a))
	vecB := mat.NewVecDense(len(b), float32To64(b))

	dot := mat.Dot(vecA, vecB)
	normA := mat.Norm(vecA, 2)
	normB := mat.Norm(vecB, 2)

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dot / (normA * normB)
}

func float32To64(s []float32) []float64 {
	d := make([]float64, len(s))
	for i, v := range s {
		d[i] = float64(v)
	}
	return d
}

func FindBestChunk(questionEmbedding []float32, chunkEmbeddings [][]float32) int {
	bestIndex := -1
	maxSimilarity := -2.0 

	for i, chunkEmb := range chunkEmbeddings {
		similarity := CosineSimilarity(questionEmbedding, chunkEmb)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
			bestIndex = i
		}
	}

	return bestIndex
}
