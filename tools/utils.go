package tools

import (
	"fmt"
	"math/big"

	"gonum.org/v1/gonum/mat"
)

// DenseToBigIntSlice 将 *mat.Dense 转换为 [][]*big.Int
func DenseToBigIntSlice(dense *mat.Dense) [][]*big.Int {
	rows, cols := dense.Dims()
	result := make([][]*big.Int, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]*big.Int, cols)
		for j := 0; j < cols; j++ {
			val := big.NewInt(int64(dense.At(i, j)))
			result[i][j] = val
		}
	}
	return result
}

func MatrixMultiplyBigInt(A, B [][]*big.Int) [][]*big.Int {
	// 获取矩阵的维度
	m := len(A)    // A 的行数
	n := len(A[0]) // A 的列数 (也是 B 的行数)
	p := len(B[0]) // B 的列数

	// 初始化结果矩阵 C，维度为 m x p
	C := make([][]*big.Int, m)
	for i := range C {
		C[i] = make([]*big.Int, p)
		for j := range C[i] {
			C[i][j] = big.NewInt(0) // 初始化每个元素为 0
		}
	}

	// 矩阵乘法计算
	for i := 0; i < m; i++ {
		for j := 0; j < p; j++ {
			sum := big.NewInt(0) // 用于存储 C[i][j]
			for k := 0; k < n; k++ {
				// 计算 A[i][k] * B[k][j]
				product := new(big.Int).Mul(A[i][k], B[k][j])
				sum.Add(sum, product) // 累加结果
			}
			C[i][j] = sum
		}
	}

	return C
}

// 获取二维矩阵指定列
func GetColumn(matrix [][]string, colIndex int) []string {
	// 创建一个切片来存储指定列的值
	column := make([]string, len(matrix))

	// 遍历矩阵的每一行，将指定列的值加入到列切片中
	for i := 0; i < len(matrix); i++ {
		column[i] = matrix[i][colIndex]
	}

	return column
}

func GetColumnFromS(matrix [][]*big.Int, colIndex int) []*big.Int {
	// 创建一个切片来存储指定列的值
	column := make([]*big.Int, len(matrix))

	// 遍历矩阵的每一行，将指定列的值加入到列切片中
	for i := 0; i < len(matrix); i++ {
		column[i] = matrix[i][colIndex]
	}

	return column
}

// 将 []*big.Int 转换为 *mat.VecDense
func BigIntSliceToVecDense(bigInts []*big.Int) (*mat.VecDense, error) {
	n := len(bigInts)
	data := make([]float64, n)

	for i, val := range bigInts {
		// 检查 big.Int 是否可以安全转换为 float64
		if !val.IsInt64() {
			return nil, fmt.Errorf("big.Int value %s is too large to convert to float64", val.String())
		}
		data[i] = float64(val.Int64())
	}

	// 使用转换后的数据构造 mat.VecDense
	return mat.NewVecDense(n, data), nil
}

// 将 []*big.Int 转换为 float64
func BigIntSliceToFloat64(bigInts []*big.Int) ([]float64, error) {
	n := len(bigInts)
	data := make([]float64, n)

	for i, val := range bigInts {
		// 检查 big.Int 是否可以安全转换为 float64
		// if !val.IsInt64() {
		// 	return nil, fmt.Errorf("big.Int value %s is too large to convert to float64", val.String())
		// }
		f := new(big.Float).SetInt(val)
		floatValue, _ := f.Float64()
		data[i] = floatValue
	}

	res := data
	// 使用转换后的数据构造 mat.VecDense
	return res, nil
}

// 求解方程式
func SolveEquation(A *mat.Dense, b *mat.VecDense) (*mat.VecDense, error) {
	// Get the dimensions of A
	r, c := A.Dims()
	if r != c {
		return nil, fmt.Errorf("matrix A must be square, got dimensions %dx%d", r, c)
	}

	// Ensure the dimensions of b match A
	if b.Len() != r {
		return nil, fmt.Errorf("vector b must have length %d to match matrix A, got %d", r, b.Len())
	}

	// Prepare a vector to store the solution
	x := mat.NewVecDense(r, nil)

	// Solve A * x = b
	err := x.SolveVec(A, b)
	if err != nil {
		return nil, fmt.Errorf("error solving equation: %v", err)
	}

	return x, nil
}
func VecDenseToBigIntSlice(vec *mat.VecDense) []*big.Int {
	// Get the raw data from VecDense
	rawData := vec.RawVector().Data

	// Prepare the output slice
	result := make([]*big.Int, len(rawData))

	// Convert each float64 to *big.Int
	for i, val := range rawData {
		bigInt := new(big.Int)
		bigInt.SetInt64(int64(val)) // Convert float64 to int64 and then to *big.Int
		result[i] = bigInt
	}

	return result
}

func BigIntMatrixToDense(matrix [][]*big.Int) *mat.Dense {
	// Get dimensions of the input matrix
	rows := len(matrix)
	if rows == 0 {
		return mat.NewDense(0, 0, nil)
	}
	cols := len(matrix[0])

	// Create a 1D slice to hold float64 data for *mat.Dense
	data := make([]float64, rows*cols)

	// Fill the data slice
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Convert *big.Int to float64
			val := new(big.Float).SetInt(matrix[i][j])
			floatVal, _ := val.Float64() // Handle overflow or precision loss if needed
			data[i*cols+j] = floatVal
		}
	}

	// Create and return *mat.Dense
	return mat.NewDense(rows, cols, data)
}

func ExtractSubmatrix(matDense *mat.Dense, b int) *mat.Dense {
	// 获取原矩阵的尺寸
	rows, cols := matDense.Dims()

	// 检查是否有足够的行
	if b > rows {
		panic("b exceeds the number of rows in the matrix")
	}

	// 创建新的子矩阵
	subMatrix := mat.NewDense(b, cols, nil)

	// 将前 b 行复制到新矩阵
	for i := 0; i < b; i++ {
		for j := 0; j < cols; j++ {
			subMatrix.Set(i, j, matDense.At(i, j))
		}
	}

	return subMatrix
}
