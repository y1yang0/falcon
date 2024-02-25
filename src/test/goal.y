func bubbleSort(arr []int, len int){
	let n= len
	for i=0;i<n-1;i+=1{
		for j=0;j<n-i-1;j+=1{
			if arr[j]>arr[j+1]{
				let t = arr[j+1]
				arr[j+1]=arr[j]
				arr[j]=t
			}
		}
	}
}

func sqrtNewton(x double, epsilon double)  {
    guess = x / 2.0
    for ((guess * guess - x) >= epsilon) || ((guess * guess - x) <= -epsilon) {
        guess = guess - (guess*guess-x)/(2*guess)
    }
    return guess
}

func main(){
    let board=[8,8]
    for i=0;i<8;i+=1{
        board[i]=i
    }
    for i=0;i<8;i+=1{
        for j=0;j<8;j+=1{
            if board[i]==board[j] || i-j==board[i]-board[j] || i-j==board[j]-board[i]{
                cprint("fail")
                return 0
            }
        }
    }
    cprint("success")
    return 0
}

	func inrange(v int, a int, b int)  {
		if v>=a && v<=b {
			return true
		}
		return false
	}
	func main(){
		assert(inrange(1,0,2),true)
		return 0
	}

func main(){
	i = 0
	while i <10 {
		cprint(i)
		i+=1

		if i==4 {
			break
		}
	}
	assert(i,4)
	return 0
}

func main(){
	let a=[9,2]
	bubbleSort(a,2)
	cprint(a[0])
	cprint(a[1])
}