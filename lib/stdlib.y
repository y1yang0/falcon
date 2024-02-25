// The Standard Library

func assert(a int,b int)
func assert_bool(a bool, b bool)

func cprint(a int)
func cprint_long(a long)
func cprint_bool(a bool)
func cprint_arr(arr []int, len int)

func sort(arr []int, len int){
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

func min(a int, b int) int{
	return a < b ? a : b
}

func max(a int, b int) int{
	return a > b ? a : b
}

func clamp(v int, min int, max int)int {
	return v < min ? min : v > max ? max : v
}