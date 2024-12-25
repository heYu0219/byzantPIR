package main

import (
	"byzantine-PIR/tools"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"gonum.org/v1/gonum/mat"
)

func main() {
	config := tools.InitConfig("config")
	// 元素大小
	num_bytes, _ := strconv.Atoi(config["num_bytes"])
	//文件数量
	m, _ := strconv.Atoi(config["m"])
	//服务器数量
	n, _ := strconv.Atoi(config["n"])
	//错误元素数量
	errorCount, _ := strconv.Atoi(config["errorCount"])
	//文件长度
	l, _ := strconv.Atoi(config["l"])
	//查询索引
	I, _ := strconv.Atoi(config["I"])
	//随机数α β
	a, _ := strconv.Atoi(config["a"])
	b, _ := strconv.Atoi(config["b"])
	b_Str := new(big.Int)
	b_bigint, _ := b_Str.SetString(config["b"], 10)

	fmt.Println("=============preprocessing==========")
	prepro_start := time.Now()

	//生成原始数据库
	X, _ := tools.GenerateRawDB(l, m, num_bytes)
	X_dense := tools.BigIntMatrixToDense(X)

	//定义编码矩阵V
	V_dense := tools.GenerateFullRankMatrix(n, l)
	//将dense矩阵转换为bigint格式
	V := tools.DenseToBigIntSlice(V_dense)
	V_1 := tools.Pseudoinverse(V_dense)

	//计算编码后的文件系统
	Y := tools.MatrixMultiplyBigInt(V, X)
	Y_dense := mat.NewDense(n, m, nil)
	Y_dense.Mul(V_dense, X_dense)

	//生成重构矩阵B
	B_dense := tools.GenerateSubMatrixFullRank(n, n)
	// B := tools.DenseToBigIntSlice(B_dense)

	//生成稀疏矩阵
	S := mat.NewDense(n, m, nil)
	S.Mul(B_dense, Y_dense)
	// S := tools.MatrixMultiplyBigInt(B, Y)
	//储存S的最后一行
	// message := S[len(S)-1]

	//生成校验矩阵
	H := tools.GenHashMat(Y)

	//分布式存储文件
	Servers := tools.DistributeFile(Y)
	prepro_end := time.Since(prepro_start)

	fmt.Println("预处理花费时间：", prepro_end)
	for i := 0; i < 10; i++ {
		fmt.Println("=============online=================")
		start := time.Now()

		query_start := time.Now()
		//构建n个查询
		Q := make([][2][]*big.Int, n) // Each element will store {q1, q2}

		// Generate queries
		for j := 0; j < n; j++ {
			q1, q2 := tools.GenQuery(a, b, m, I)
			Q[j] = [2][]*big.Int{q1, q2} // Assign q1 and q2 which are of type []*big.Int
		}
		query_end := time.Since(query_start)
		fmt.Println("构建查询花费时间：", query_end)

		resp_start := time.Now()
		//服务器生成响应
		A := make([][2]*big.Int, n)

		// 遍历每一对 Servers 和 Q，调用 GenRespond 函数
		for i := 0; i < n; i++ {
			A[i][0], A[i][1] = tools.GenRespond(Servers[i], Q[i])
		}
		resp_end := time.Since(resp_start)
		fmt.Println("服务器响应时间：", resp_end)

		//========客户端处理服务器响应
		// Respond 是一个 []*big.Int 类型的切片
		Respond := make([]*big.Int, 0, n)

		// 假设 A 是一个 [][]*big.Int 类型的切片，存储每个服务器的响应数据
		for i := 0; i < n; i++ {
			// 调用 CalRespond 函数并将结果添加到 Respond 切片中
			res := tools.CalRespond(b_bigint, A[i])
			Respond = append(Respond, res) // Append the result to Respond slice
		}

		//指定错误服务器
		specifiedByzant := make([]int, errorCount)
		for i := 0; i < errorCount; i++ {
			specifiedByzant[i] = i
		}
		// errorCountLen := len(specifiedByzant)

		//构造错误
		erroResponse := tools.AddError(specifiedByzant, Respond)

		//验证
		H_I := tools.GetColumn(H, I)
		_, ByzantServer, HonestServers := tools.Verify(erroResponse, H_I)

		//恢复结果
		// correct := erroResponse

		S_column := mat.Col(nil, I, S)
		// S_column := tools.GetColumnFromS(S, I)
		// S_I, _ := tools.BigIntSliceToVecDense(S_column)
		S_I := mat.NewVecDense(len(S_column), S_column)
		erroResponse_float, _ := tools.BigIntSliceToFloat64(erroResponse)

		b_error := mat.NewVecDense(errorCount, nil)
		if errorCount != 0 {
			if errorCount < n {
				b_error = tools.Reconstruct(B_dense, S_I, erroResponse_float, ByzantServer, HonestServers, errorCount)
				fmt.Println("小于n个错误：", b_error)
			} else {

				correct, _ := tools.SolveEquation(B_dense, S_I)
				fmt.Println("等于n个错误：", correct)
			}
		}
		correct_b := tools.VecDenseToBigIntSlice(b_error)

		//重组
		finalRes := make([]*big.Int, n)
		for i := 0; i < n; i++ {
			finalRes[i] = big.NewInt(0) // 初始化为 0
		}
		if errorCount != 0 {
			// 填充 Byzantine 服务器的修正结果
			for i, server := range ByzantServer {
				roundedCorrect := new(big.Int)
				roundedCorrect = correct_b[i]
				// b_error[i].Int(roundedCorrect) // 将修正结果四舍五入为整数
				finalRes[server] = roundedCorrect
			}

			// 填充 Honest 服务器的响应值
			for _, server := range HonestServers {
				finalRes[server] = Respond[server]
			}
		} else {
			// 如果没有错误服务器，直接使用 Respond 作为结果
			finalRes = Respond
		}

		//解码文件
		//todo 传入的y是错误的y
		y, _ := tools.BigIntSliceToFloat64(finalRes)
		result, _ := tools.Decode(V_1, y)
		fmt.Println("解码后的文件维度:", result.Cap())
		// fmt.Println("原始文件:", tools.GetColumnFromS(X, I))

		end := time.Since(start)
		fmt.Println("total time: ", end)
	}

}
