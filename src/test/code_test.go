// Copyright (c) 2024 The Falcon Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package test

import (
	"falcon/compile"
	"falcon/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ExecExpect(source string, expect ...string) {
	// Compile
	app := compile.CompileText(source)
	wd := filepath.Dir(app)
	// Execute
	output := utils.ExecuteCmd(wd, app)
	os.Remove(app)
	for _, str := range expect {
		if !strings.Contains(output, str) {
			fmt.Printf(("== Source:\n%v== Output:\n%v\n== Expect:\n%v\nFailed\n"),
				source, output, expect)
			panic("fail")
			// os.Exit(1)
		}
	}
}

func TestApp1(t *testing.T) {
	source := `
	func foo(){
		for i=0;i<3;i+=1{
			cprint(i)
		}
	}
	func main(){
		foo()
	}
	`
	ExecExpect(source, "0", "1", "2")

	source = `
	func foo(){
		if false {
			cprint(1024)
		}
	}
	func main(){
		foo()
	}
	`
	ExecExpect(source)

	source = `
	func check(v int) {
		let r  =6
		if v==1 {
			r=1
		} else {
			if v==2 {
				r=2
			} else {
				if v==3 {
					r=3
				} else {
					if v==4 {
						r=4
					} else {
						if v==5 {
							r=5
						} else {
							r=6
						}
					}
				}
			}
		}
		cprint(r)
	}

	func main(){
		check(1)
		check(2)
		check(3)
		check(4)
		check(5)
		check(6)
	}
	`
	ExecExpect(source, "1", "2", "3", "4", "5", "6")

	source = `
	func main(){
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		let sum=0
		for i=1;i<=100;i+=1{
			sum = sum + i
		}
		assert(sum, 5050)
		cprint(sum)
	}
	`
	ExecExpect(source)

	source = `
	func foo() int {
		return 1024
	}
	func main(){
		cprint(foo())
	}
	`
	ExecExpect(source, "1024")

	source = `
	func foo() int {
		return 1024
	}
	func main(){
		assert(foo(),1024)
	}
	`
	ExecExpect(source)
}

func TestApp2(t *testing.T) {
	source := `
	func main(){
		let arr=[0,1,0,1]
		assert(arr[0],arr[2])
		assert(arr[1],arr[3])
		let sum = arr[0]+arr[1]+arr[2]+arr[3]
		assert(sum,2)
		if arr[0]==0 {
			arr[0] = 1
		}
		assert(arr[0],1)
	}
	`
	ExecExpect(source)
}

func TestApp3(t *testing.T) {
	source := `
	func main(){
		let arr=[0,0,0,0]
		for i=0;i<4;i+=1{
			arr[i]=i
		}
		assert(arr[0],0)
		assert(arr[1],1)
		assert(arr[2],2)
		assert(arr[3],3)
	}
	`
	ExecExpect(source)
}

func TestApp4(t *testing.T) {
	source := `
	func main(){
		for i=1; i<=1; i+=1{
			for k=3; k<=4; k+=1{
				cprint(i+k)
			}
		}
	}
	`
	ExecExpect(source, "4", "5")
}

func TestArithmetic(t *testing.T) {
	source := `
	func main(){
		let p = 3+4
		let q = p+4
		assert(p, 7)
		assert(q, 11)
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		let p = 3-4
		let q = p - 5
		assert(p,-1)
		assert(q,-6)
		let t =  q + 6
		assert(t,0)
	}
	`
	ExecExpect(source)
}

func TestFibonacci(t *testing.T) {
	source := `
	func fibo(n int) int {
		if n==0{
			return 0
		}
		if n==1{
			return 1
		}
		return fibo(n-1)+fibo(n-2)
	}

	func main(){
		assert(fibo(0),0)
		assert(fibo(1),1)
		assert(fibo(2), 1)
		assert(fibo(3), 2)
		assert(fibo(4), 3)
		assert(fibo(5), 5)
		assert(fibo(6), 8)
		assert(fibo(7), 13)
		assert(fibo(8), 21)
		assert(fibo(9), 34)
		assert(fibo(10),55)
		assert(fibo(20), 6765)
	}
	`
	ExecExpect(source)
}

func TestSort(t *testing.T) {
	source := `
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
	func main(){
		let a=[2,9,3,7,0,-10,8,-6]
		bubbleSort(a,8)
		assert(a[7],9)
		assert(a[6],8)
		assert(a[5],7)
		assert(a[4],3)
		assert(a[3],2)
		assert(a[2],0)
		assert(a[1],-6)
		assert(a[0],-10)
	}
	`
	ExecExpect(source)
}

func TestConditionalExpr(t *testing.T) {
	source := `
	func f(v int) int {
        return v < 100 ? 666 : 555
	}
	func main(){
		assert(f(50),666)
		assert(f(200),555)
		assert(f(100),555)
		assert(f(99),666)
	}
	`
	ExecExpect(source)
}

func TestCmp(t *testing.T) {
	source := `
	func f1(v int) int {
        if v < 100 {
			return 555
		} else {
			return 666
		} 
	}
	func f2(v int) int {
        if v <= 100 {
			return 555
		} else {
			return 666
		} 
	}
	func f3(v int) int {
        if v > 100 {
			return 555
		} else {
			return 666
		} 
	}
	func f4(v int) int {
        if v >= 100 {
			return 555
		} else {
			return 666
		} 
	}
	func f5(v int) int {
        if v == 100 {
			return 555
		} else {
			return 666
		} 
	}
	func f6(v int) int {
        if v != 100 {
			return 555
		} else {
			return 666
		} 
	}
	func main(){
		assert(f1(50),555)
		assert(f1(200),666)
		assert(f2(100),555)
		assert(f2(101),666)
		assert(f3(101),555)
		assert(f3(100),666)
		assert(f4(100),555)
		assert(f4(99),666)
		assert(f5(100),555)
		assert(f5(101),666)
		assert(f6(101),555)
		assert(f6(100),666)
	}
	`
	ExecExpect(source)
}

func TestArray(t *testing.T) {
	source := `
	func main(){
		let a=[1,2,3]
		assert(a[0],1)
		assert(a[1],2)
		assert(a[2],3)
		a[0]=4
		assert(a[0],4)
	}
	`
	ExecExpect(source)
}

func TestLet(t *testing.T) {
	source := `
	func main(){
		let a=1
		assert(a,1)
		a=2
		assert(a,2)
	}
	`
	ExecExpect(source)
}

func TestIf(t *testing.T) {
	source := `
	func main(){
		if true {
			cprint(1)
		}
		if false {
			cprint(2)
		}
		if 1==1 {
			cprint(3)
		}
		if 1!=1 {
			cprint(4)
		}
	}
	`
	ExecExpect(source, "1", "3")
}

func TestLogicalOp(t *testing.T) {
	source := `
	func main(){
		if true && true {
			cprint(1)
		}
		if false || true {
			cprint(2)
		}
		if false && true {
			cprint(3)
		}
		if false || false {
			cprint(4)
		}
	}
	`
	ExecExpect(source, "1", "2")
}

func TestLogicalCompare(t *testing.T) {
	source := `
	func main(){
		if 1==1 && 2==2 {
			cprint(1)
		}
		if 1==1 || 2==3 {
			cprint(2)
		}
		if 1==2 && 2==2 {
			cprint(3)
		}
		if 1==2 || 2==3 {
			cprint(4)
		}
	}
	`
	ExecExpect(source, "1", "2")
}

func TestComplexLogical(t *testing.T) {
	source := `
	func main(){
		if 1==1 && (2==2 || 3==3) {
			cprint(1)
		}
		if 1==1 && (2==3 || 3==3) {
			cprint(2)
		}
		if 1==2 && (2==2 || 3==3) {
			cprint(3)
		}
		if 1==2 && (2==3 || 3==3) {
			cprint(4)
		}
	}
	`
	ExecExpect(source, "1", "2")
}

func TestIntInRange(t *testing.T) {
	source := `
	func inrange(v int, a int, b int) bool {
		if v>=a && v<=b {
			return true
		}
		return false
	}
	func main(){
		assert_bool(inrange(1,1,1),true)
		assert_bool(inrange(1,1,2),true)
		assert_bool(inrange(1,2,2),false)
		assert_bool(inrange(1,2,3),false)
		assert_bool(inrange(1,0,0),false)
		assert_bool(inrange(1,0,1),true)
		assert_bool(inrange(1,0,2),true)
	}
	`
	ExecExpect(source)
}

func TestFindElem(t *testing.T) {
	source := `
	func findElem(arr []int, len int, v int) int {
		for i=0;i<len;i+=1{
			if arr[i]==v {
				return i
			}
		}
		return -1
	}
	func findExist(arr []int, len int, v int) bool {
		for i=0;i<len;i+=1{
			if arr[i]==v {
				return true
			}
		}
		return false
	}
	func main(){
		let a=[1,2,3,4,5]
		assert(findElem(a,5,0),-1)
		assert(findElem(a,5,1),0)
		assert(findElem(a,5,2),1)
		assert(findElem(a,5,3),2)
		assert(findElem(a,5,4),3)
		assert(findElem(a,5,5),4)
		assert(findElem(a,5,6),-1)
		assert_bool(findExist(a,5,0),false)
		assert_bool(findExist(a,5,1),true)
		assert_bool(findExist(a,5,2),true)
		assert_bool(findExist(a,5,3),true)
		assert_bool(findExist(a,5,4),true)
		assert_bool(findExist(a,5,5),true)
		assert_bool(findExist(a,5,6),false)
	}
	`
	ExecExpect(source)
}

func TestBasicArithmetic(t *testing.T) {
	source := `
	func main(){
		assert(1+1,2)
		assert(1-1,0)
		assert(1*1,1)
		assert(1/1,1)
		assert(4/3,1)
		assert(0/1,0)
		//assert(1/0,0)
		//assert(1%1,0)
		assert(1+2*3,7)
		assert((1+2)*3,9)
		assert(1+(2*3),7)
		assert(1-(2*3),-5)
		assert(1-(2+3),-4)
		assert(1-(2+3)*4,-19)
		assert((1-(2+3))*4,-16)
		assert(1-(2+3)*4/2,-9)
		assert(-16*2,-32)
		assert(16*-2,-32)
		assert(16/-2, -8)
		assert(-16/2,-8)
		assert((1-(2+3))*4/2,-8)
		assert(1*2*3*4*5*6*7*8*9*10,3628800)
		assert(3628800/10/9/8/7/6/5/4/3/2,1)
		assert(-1/-1,1)
		assert(-1/1,-1)
		assert(1/-1,-1)
		assert(-1*-1,1)
		assert(-1*1,-1)
		assert(1*-1,-1)
		assert(0*0,0)
		assert(0*1,0)
		assert(1*0,0)
		assert(0*0*0*0*0*0*0*0*0*0,0)
		assert(0*0*0*0*0*0*0*0*0*1,0)
	}
	`
	ExecExpect(source)
}

func TestWhile(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<3 {
			cprint(i)
			i+=1
		}

		i=0
		while i<3 {
			i+=1
		}
		assert(i,3)

		// mix while and for
		i=0
		for j=0;j<3;j+=1{
			i+=1
			while i<3 {
				i+=1
			}
		}
		assert(i,5)
	}
	`
	ExecExpect(source, "0", "1", "2")
}

func TestBreak(t *testing.T) {
	source := `
	func main(){
		while true {
			break
		}
		while false {
			break
		}
		while true {
			if true{
				break
			}
		}
		while true {
			if false{
				break
			}else {
				break
			}
		}
		while true {
			while true {
				break
			}
			break
		}
		while true {
			while false {
				break
			}
			break
		}
		while true {
			while true {
				if true {
					break
				}
			}
			break
		}
		while true {
			if true {
				if true {
					if true {
						break
					}
				}
			}
			break
		}
		while true {
			while true {
				while true {
					break
				}
				break
			}
			break
		}
	}
	`
	ExecExpect(source)
}

func TestDeadBreak(t *testing.T) {
	source := `
	func main(){
		while true {
			break
			// dead then...
			let p = 3
			assert(p,1024)
		}
		while true {
			let p = 10
			while true {
				break
				p += 3
			}
			assert(p,10)
			break
		}
		while true {
			break
			break
			break
		}
	}
	`
	ExecExpect(source)
}

func TestBreak2(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<10 {
			if i==2 {
				break
			}
			i+=1
		}
		assert(i,2)
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		let i=0
		while i<10 {
			if i==2 {
				break
			}
			i+=1
			if i==3 {
				break
			}
		}
		assert(i,2)
	}
	`
	ExecExpect(source)
}

func TestContinue(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<3 {
			i+=1
			continue
		}
		assert(i,3)
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		let i=0
		while i<3 {
			i+=1
			if i==2 {
				continue
			}
		}
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDeadContinue(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<3 {
			i+=1
			continue
			// dead then...
			let p = 3
			assert(p,1024)
		}
	}
	`
	ExecExpect(source)
}

func TestDeadContinue2(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<3 {
			i+=1
			if i==2 {
				continue
				// dead then...
				let p = 3
				assert(p,1024)
			}
		}
	}
	`
	ExecExpect(source)
}

func TestContinueAndBreak(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<10 {
			if i <5 {
				i +=1
				continue
			}
			i+=2
			if i==7 {
				break
			}
		}
		assert(i,7)
	}
	`
	ExecExpect(source)
}

func TestBreakAndContinue(t *testing.T) {
	source := `
	func main(){
		i=0
		while true {
			if i==3{
				break
			}
			i+=1
		}
		assert(i,3)
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		i=0
		while true {
			if i<3{
				i+=1
				continue
			} else {
				break
			}
		}
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestIfElseIf(t *testing.T) {
	source := `
	func main(){
		if true {
			cprint(1)
		} else if false {
			cprint(2)
		} else if true {
			cprint(3)
		} else {
			cprint(4)
		}
	}
	`
	ExecExpect(source, "1")
}

func TestIfElseIf2(t *testing.T) {
	source := `
	func main(){
		let i =3
		if i==4 {
			cprint(4)
		}else if i==5{
			cprint(5)
		}else if i==6 {
			cprint(6)
		}else {
			cprint(55)
		}
	}
	`
	ExecExpect(source, "55")
}

func TestIfElseIfInLoop(t *testing.T) {
	source := `
	func main(){
		for i=0;i<5;i+=1{
			if i==0 {
				cprint(0)
			}else if i==1 {
				cprint(1)
			}else if i==2 {
				cprint(2)
			}else if i==3 {
				cprint(3)
			}else {
				cprint(4)
			}
		}
	}
	`
	ExecExpect(source, "0", "1", "2", "3", "4")
}

func TestIfElseIfElseWithBreak(t *testing.T) {
	source := `
	func main(){
		let i = 0
		while true {
			if i==0 {
				break
			}else if i==1 {
				break
			}else {
				break
			}
		}
	}
	`
	ExecExpect(source)

	source = `
	func main(){
		let i = 100
		while true {
			if i==0 {
				break
			}else if i==1 {
				break
			}
			i+=1
			break
		}
		assert(i,101)
	}
	`
	ExecExpect(source)
}

func TestArithmeticTypes(t *testing.T) {
	source := `
	func main(){
		let p long = 333
	}
	`
	ExecExpect(source)
}

func TestExoticIf(t *testing.T) {
	source := `
	func main(){
		let p = 2
		if p == 3 {
			cprint(0)
		} else while p < 10 {
			p = p + 1
			cprint(p)
		}

		p = 1
		if p == 3 {
			cprint(0)
		} else for i=0;i<10;i+=1{
			p = p + 1
			cprint(p)
		}
	}
	`
	ExecExpect(source, "3", "4", "5", "6", "7", "8", "9", "10", "11")
}

func TestLetWithoutInit(t *testing.T) {
	source := `
	func main(){
		let p int
		assert(p,0)
	}
	`
	ExecExpect(source)
}

func TestLogicalNot(t *testing.T) {
	source := `
	func main(){
		assert_bool(!true,false)
		assert_bool(!false,true)
	}
	`
	ExecExpect(source)
}

func TestLogicalNot2(t *testing.T) {
	source := `
	func main(){
		let p = true
		if !p {
			cprint(0)
		} else {
			cprint(1)
		}
	}
	`
	ExecExpect(source, "1")
}

func TestWhileWithIf(t *testing.T) {
	source := `
	func main(){
		let i=0
		while i<10 {
			if i==5 {
				break
			}
			i+=1
		}
		assert(i,5)
	}
	`
	ExecExpect(source)
}

func TestArithmeticAndAssign(t *testing.T) {
	source := `
	func main(){
		let p = 1
		p+=1
		assert(p,2)
		p-=1
		assert(p,1)
		p*=2
		assert(p,2)
		p/=2
		assert(p,1)
		p%=1
		assert(p,0)
		p++
		assert(p,1)
		p--
		assert(p,0)
	}
	`
	ExecExpect(source)
}

func TestArithmeticAndAssignBit(t *testing.T) {
	source := `
	func main(){
		let p = 1
		p<<=1
		assert(p,2)
		p>>=1
		assert(p,1)
	}
	`
	ExecExpect(source)
}

func TestArithmeticAndAssignBit2(t *testing.T) {
	source := `
	func main(){
		let p = 1
		p&=1
		assert(p,1)
		p|=1
		assert(p,1)
		p^=1
		assert(p,0)
	}
	`
	ExecExpect(source)
}

func TestMod(t *testing.T) {
	source := `
	func main(){
		assert(4%3,1)
		assert(4%-3,1)
		assert(-4%3,-1)
		assert(-4%-3,-1)
	}
	`
	ExecExpect(source)
}

func TestStringBinaryEx(t *testing.T) {
	source := `
	func main(){
		let a = "aa"
		let b = "bb"
		let c = a + b
		assert_string(c, "aabb")
		let d = a == b
		assert_bool(d, false)
		let e = a == "aa"
		assert_bool(e, true)
		let f = a != b
		assert_bool(f, true)
		let g = a != "aa"
		assert_bool(g, false)
		let h = a < b
		assert_bool(h, true)
		let i = a > b
		assert_bool(i, false)
		let j = a <= b
		assert_bool(j, true)
		let k = a >= b
		assert_bool(k, false)
	}
	`
	ExecExpect(source)
}

func TestHelloWorld(t *testing.T) {
	source := `
	func main(){
		let a = "Hello,"
		let b = " World!"
		let c = a + b
		cprint_string(c)
	}
	`
	ExecExpect(source, "Hello, World!")
}

func TestString(t *testing.T) {
	source := `
	func main(){
		let a = "abc"
		let x1 = a[0]
		let x2 = a[1]
		let x3 = a[2]
		assert_char(x1, 'a')
		assert_char(x2, 'b')
		assert_char(x3, 'c')
		assert_char('a', a[0])
		assert_char('b', a[1])
		assert_char('c', a[2])
	}
	`
	ExecExpect(source)
}

func TestDoWhile(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDoWhile2(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			if i==2 {
				continue
			}
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDoWhile3(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			if i==2 {
				break
			}
		} while i<3
		assert(i,2)
	}
	`
	ExecExpect(source)
}

func TestDoWhileNesting(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			let j=0
			do {
				j+=1
				if j==2 {
					break
				}
			} while j<3
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDoWhileNesting2(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			let j=0
			do {
				j+=1
				if j==2 {
					continue
				}
			} while j<3
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDoWhileNestingWhile(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			let j=0
			while j<3 {
				j+=1
				if j==2 {
					break
				}
			}
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestDoWhileNestingWhile2(t *testing.T) {
	source := `
	func main(){
		let i=0
		do {
			i+=1
			let j=0
			while j<3 {
				j+=1
			}
		} while i<3
		assert(i,3)
	}
	`
	ExecExpect(source)
}

// test for and do while and while
func TestLoopRemix(t *testing.T) {
	source := `
	func main(){
		let i=0
		for j=0;j<3;j+=1{
			i+=1
			let k=0
			while k<3 {
				k+=1
			}
		}
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestLoopRemix2(t *testing.T) {
	source := `
	func main(){
		let i=0
		for j=0;j<3;j+=1{
			i+=1
			let k=0
			do {
				k+=1
			} while k<3
		}
		assert(i,3)
	}
	`
	ExecExpect(source)
}

func TestLoopRemix3(t *testing.T) {
	source := `
	func main(){
		let i=0
		let j=0
		while j<3 {
			j+=1
			let k=0
			do {
				k+=1
			} while k<3
			i+=1
		}
		assert(i,3)
	}
	`
	ExecExpect(source)
}

// test loop with return
func TestLoopReturn(t *testing.T) {
	source := `
	func main(){
		let i=0
		for j=0;j<3;j+=1{
			i+=1
			if j==1 {
				return
			}
		}
		assert(1,2)
	}
	`
	ExecExpect(source)
}

// func TestSubstring(t *testing.T) {
// 	source := `
// 	func substring(s string, l int, r int) string {
// 		let n=r-l+1
// 		let res=""
// 		for i=0;i<n;i+=1{
// 			res=res +s[l+i]
// 		}
// 		return res
// 	}
// 	func main(){
// 		assert_string(substring("hello",0,4),"hello")
// 		assert_string(substring("hello",0,0),"h")
// 		assert_string(substring("hello",1,2),"el")
// 	}
// 	`
// 	ExecExpect(source)
// }
