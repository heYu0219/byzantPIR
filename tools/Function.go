package tools

import (
	cRand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	mRand "math/rand"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"gonum.org/v1/gonum/mat"
	// "gonum.org/v1/gonum/lapack"
	"bufio"
	"io"
	"os"
	"strings"
)

// GenerateInteger 生成指定字节大小的整数
func GenerateInteger(bitSize int) *big.Int {
	// Generate a random big.Int with the specified bit size
	max := new(big.Int).Lsh(big.NewInt(1), uint(bitSize)) // 2^bitSize
	n, err := cRand.Int(cRand.Reader, max)
	if err != nil {
		panic(fmt.Sprintf("Error generating random big.Int: %v", err))
	}
	return n
}

// GenerateRawDB 生成 m x n 的随机整数数据库
func GenerateRawDB(m, n, numBytes int) ([][]*big.Int, error) {
	// 创建 m x n 的二维切片
	rawDB := make([][]*big.Int, m)
	for i := 0; i < m; i++ {
		rawDB[i] = make([]*big.Int, n)
		for j := 0; j < n; j++ {
			element := GenerateInteger(numBytes)

			rawDB[i][j] = element
		}
	}
	return rawDB, nil
}

// 生成范德蒙矩阵 B矩阵 任意子矩阵满秩
// GenerateSubMatrixFullRank generates a Vandermonde matrix where any submatrix is full rank
func GenerateSubMatrixFullRank(m, n int) *mat.Dense {
	if m > n {
		panic("Matrix rows (m) should be less than or equal to columns (n) to ensure full rank")
	}

	// Initialize random seed
	mRand.Seed(time.Now().UnixNano())

	// Generate distinct alpha values
	alpha := make([]float64, m)
	for i := 0; i < m; i++ {
		alpha[i] = float64(mRand.Intn(100) + 1) // Ensure non-zero distinct values
	}

	// Create Vandermonde matrix
	data := make([]float64, m*n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			data[i*n+j] = pow(alpha[i], j) // Alpha_i^j
		}
	}

	return mat.NewDense(m, n, data)
}

// pow computes the power of a base (x^y)
func pow(base float64, exp int) float64 {
	result := 1.0
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}

// 生成m*l的满秩矩阵 V
// GenerateFullRankMatrix generates an m x n random full-rank matrix.
func GenerateFullRankMatrix(m, n int) *mat.Dense {
	mRand.Seed(time.Now().UnixNano())
	for {
		// Generate a random m x n matrix
		data := make([]float64, m*n)
		for i := range data {
			data[i] = float64(mRand.Intn(100) + 1) // Random values between 1 and 100
		}
		matrix := mat.NewDense(m, n, data)

		// Check if the matrix is full rank
		if isFullRank(matrix, min(m, n)) {
			return matrix
		}
	}
}

// isFullRank checks if a matrix is full rank using SVD.
func isFullRank(matrix *mat.Dense, rank int) bool {
	var svd mat.SVD
	if ok := svd.Factorize(matrix, mat.SVDThin); !ok {
		panic("SVD factorization failed")
	}

	// Get the singular values
	singularValues := svd.Values(nil)
	tolerance := 1e-10

	// Count the number of singular values greater than the tolerance
	count := 0
	for _, value := range singularValues {
		if math.Abs(value) > tolerance {
			count++
		}
	}

	return count == rank
}

// Helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 分布式存储
func DistributeFile(Y [][]*big.Int) [][]*big.Int {
	n := len(Y) // Y 的行数
	var Servers [][]*big.Int

	// 遍历每行，将每行添加到 Servers 中
	for i := 0; i < n; i++ {
		// 创建新的切片来保存 Y[i, :]
		rowCopy := make([]*big.Int, len(Y[i]))
		copy(rowCopy, Y[i])
		Servers = append(Servers, rowCopy)
	}

	return Servers
}

// HashInteger calculates the SHA-256 hash of a given integer.
func Gethash(value *big.Int) string {
	// Convert the *big.Int value to a byte slice
	bytes := value.Bytes()

	// Compute the hash
	hasher := sha256.New()
	hasher.Write(bytes)

	// Return the hex-encoded hash
	return hex.EncodeToString(hasher.Sum(nil))
}

// 为列向量求哈希值
func GenHashVec(Y []*big.Int) []string {
	n := len(Y)
	H := make([]string, n)
	for i := 0; i < n; i++ {
		H[i] = Gethash(Y[i])
	}
	return H
}

// 为矩阵求哈希
// GenHashMat generates a hash matrix for a given 2D slice of integers.
func GenHashMat(Y [][]*big.Int) [][]string {
	n := len(Y)
	if n == 0 {
		return nil
	}
	m := len(Y[0])
	H := make([][]string, n)
	for i := 0; i < n; i++ {
		H[i] = make([]string, m)
		for j := 0; j < m; j++ {
			H[i][j] = Gethash(Y[i][j])
		}
	}
	return H
}

// 构建查询
func GenQuery(a, b, m, i int) ([]*big.Int, []*big.Int) {
	// Initialize the vector `e` with all zeros and set `e[i] = 1`
	e := make([]*big.Int, m)
	for j := 0; j < m; j++ {
		e[j] = big.NewInt(0)
	}
	e[i] = big.NewInt(1)

	// Generate the random vector `r` with integers in the range [0, 10)
	r := make([]*big.Int, m)
	for j := 0; j < m; j++ {
		r[j] = big.NewInt(int64(mRand.Intn(10))) // Generate random numbers between 0 and 9
	}

	// Calculate q1 = a * r
	q1 := make([]*big.Int, m)
	for j := 0; j < m; j++ {
		q1[j] = new(big.Int).Mul(big.NewInt(int64(a)), r[j])
	}

	// Calculate q2 = q1 + b * e
	q2 := make([]*big.Int, m)
	for j := 0; j < m; j++ {
		q2[j] = new(big.Int).Add(q1[j], new(big.Int).Mul(big.NewInt(int64(b)), e[j]))
	}

	return q1, q2
}

// 计算服务器响应
func GenRespond(Y []*big.Int, Q [2][]*big.Int) (*big.Int, *big.Int) {
	// 使用 big.Int 来进行内积计算
	a1 := new(big.Int)
	a2 := new(big.Int)

	// 计算 a1 = Y 内积 Q[0]
	for i := 0; i < len(Y); i++ {
		a1.Add(a1, new(big.Int).Mul(Y[i], Q[0][i]))
	}

	// 计算 a2 = Y 内积 Q[1]
	for i := 0; i < len(Y); i++ {
		a2.Add(a2, new(big.Int).Mul(Y[i], Q[1][i]))
	}

	return a1, a2
}

// 计算各服务器的响应值
func CalRespond(b *big.Int, A [2]*big.Int) *big.Int {
	a1 := A[0]
	a2 := A[1]

	// 计算 (a2 - a1) / b，注意要检查b是否为零
	result := new(big.Int)
	diff := new(big.Int).Sub(a2, a1) // a2 - a1
	return result.Div(diff, b)       // (a2 - a1) / b
}

// 验证服务器响应与哈希值
func Verify(Y []*big.Int, H []string) (int, []int, []int) {
	count := 0
	ByzantServer := []int{}
	HonestServer := []int{}
	n := len(Y)
	yHash := GenHashVec(Y)

	for i := 0; i < n; i++ {
		if yHash[i] != H[i] {
			count++
			ByzantServer = append(ByzantServer, i)
		} else {
			HonestServer = append(HonestServer, i)
		}
	}

	return count, ByzantServer, HonestServer
}

// Reconstruct 恢复错误响应值
func Reconstruct(B *mat.Dense, S *mat.VecDense, Respond []float64, ByzantineServers, HonestServers []int, errorCount int) *mat.VecDense {
	// b := len(ByzantineServers)

	// 提取矩阵 B 的子矩阵（选定的 b 行和对应列）
	// B_I := extractRows(B, ByzantineServers)
	B_I := ExtractSubmatrix(B, errorCount)
	B_unknown := extractColumns(B_I, ByzantineServers)
	B_known := extractColumns(B_I, HonestServers)

	// 初始化 S_I（拜占庭服务器的响应向量）
	S_I := extractRowsVec(S, ByzantineServers)

	// 减去已知响应值对 S_I 的影响
	for k, h := range HonestServers {
		col := mat.Col(nil, k, B_known)
		for i := 0; i < S_I.Len(); i++ {
			a := S_I.AtVec(i)
			rh := Respond[h]
			c := col[i]
			S_I.SetVec(i, a-rh*c)
		}
	}

	// 使用 Solve 解线性方程组 B_unknown * x = S_I
	x := mat.NewVecDense(S_I.Len(), nil)
	err := x.SolveVec(B_unknown, S_I)
	xx := x
	if err != nil {
		panic(fmt.Sprintf("Error solving system: %v", err))
	}

	return xx
}

// Helper: 提取特定列
func extractColumns(matrix *mat.Dense, columns []int) *mat.Dense {
	rows, _ := matrix.Dims()
	result := mat.NewDense(rows, len(columns), nil)
	for i, col := range columns {
		result.SetCol(i, mat.Col(nil, col, matrix))
	}
	return result
}

// Helper: 提取特定行
func extractRows(matrix *mat.Dense, rows []int) *mat.Dense {
	_, cols := matrix.Dims()
	result := mat.NewDense(len(rows), cols, nil)
	for i, row := range rows {
		result.SetRow(i, mat.Row(nil, row, matrix))
	}
	return result
}

// Helper: 提取特定行（向量版本）
func extractRowsVec(vector *mat.VecDense, rows []int) *mat.VecDense {
	result := make([]float64, len(rows))
	for i, row := range rows {
		result[i] = vector.AtVec(row)
	}
	return mat.NewVecDense(len(rows), result)
}

// 求广义逆矩阵
func Pseudoinverse(A *mat.Dense) *mat.Dense {
	// 执行 SVD 分解 A = U * Σ * Vᵀ
	var svd mat.SVD
	ok := svd.Factorize(A, mat.SVDThin)
	if !ok {
		panic("SVD factorization failed")
	}

	// 获取 U, Σ 和 V
	var U, V mat.Dense
	svd.UTo(&U)
	svd.VTo(&V)
	singularValues := svd.Values(nil)

	// 构造 Σ⁺ (Σ 的伪逆)
	m, n := len(singularValues), len(singularValues)
	SigmaPlus := mat.NewDense(n, m, nil)
	tolerance := 1e-10 // 控制零值的阈值
	for i := 0; i < len(singularValues); i++ {
		if singularValues[i] > tolerance {
			SigmaPlus.Set(i, i, 1/singularValues[i])
		}
	}

	// 计算广义逆 A⁺ = V * Σ⁺ * Uᵀ
	var VT mat.Dense
	VT.Mul(&V, SigmaPlus)
	var PseudoInv mat.Dense
	PseudoInv.Mul(&VT, U.T())

	return &PseudoInv
}

// 文件解码
func Decode(V_1 *mat.Dense, y []float64) (*mat.VecDense, error) {
	// 检查 V_1 和 y 的维度是否匹配
	// _, cols := V_1.Dims()
	// if len(y) != cols {
	// 	return nil, fmt.Errorf("dimension mismatch: matrix columns %d, vector size %d", cols, len(y))
	// }

	// 创建 y 的 VecDense 表示
	yVec := mat.NewVecDense(len(y), y)

	// 计算矩阵与向量的乘积
	result := mat.NewVecDense(V_1.RawMatrix().Rows, nil)
	result.MulVec(V_1, yVec)

	return result, nil
}

// AddError 函数：将拜占庭服务器索引对应位置的值设置为 -1
func AddError(Byzant []int, Y []*big.Int) []*big.Int {
	// Loop over the Byzantine indices and set corresponding Y values to -1
	for _, index := range Byzant {
		if index >= 0 && index < len(Y) {
			// Set the value at index to -1
			Y[index] = big.NewInt(-1)
		}
	}
	return Y
}

// 读取key=value类型的配置文件
func InitConfig(path string) map[string]string {
	config := make(map[string]string)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config
}
