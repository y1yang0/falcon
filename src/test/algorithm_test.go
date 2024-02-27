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

import "testing"

func TestJosephus(t *testing.T) {
	source := `
	func josephus(n int, k int) int {
		if n==1 {
			return 0
		}
		return (josephus(n-1,k)+k)%n
	}
	func main(){
		assert(josephus(5,3),3)
		assert(josephus(5,2),2)
		assert(josephus(5,1),4)
		assert(josephus(5,4),0)
		assert(josephus(5,5),1)
		assert(josephus(5,6),3)
	}
	`
	ExecExpect(source)
}

func TestPrimeByEratosthen(t *testing.T) {
	source := `
	func main(){
		let n = 10
		let prime = [true,true,false,false,false,false,false,false,false,false,false]
		for i=2;i*i<=n;i+=1{
			if prime[i] {
				continue
			}
			for j=i*i;j<=n;j+=i{
				prime[j]=true
			}
		}
		let cnt=0
		for i=2;i<=n;i+=1{
			if prime[i]==false {
				cnt+=1
			}
		}
		assert(cnt,4)
	}
	`
	ExecExpect(source)
}

func TestIsPrime(t *testing.T) {
	source := `
	func isPrime(n int) bool {
		if n<2 {
			return false
		}
		for i=2;i*i<=n;i+=1{
			if n%i==0 {
				return false
			}
		}
		return true
	}
	func main(){
		assert_bool(isPrime(1),false)
		assert_bool(isPrime(2),true)
		assert_bool(isPrime(3),true)
		assert_bool(isPrime(4),false)
		assert_bool(isPrime(5),true)
		assert_bool(isPrime(6),false)
		assert_bool(isPrime(7),true)
		assert_bool(isPrime(8),false)
		assert_bool(isPrime(9),false)
		assert_bool(isPrime(10),false)
	}
	`
	ExecExpect(source)
}

func TestGCD(t *testing.T) {
	source := `
	func gcd(a int, b int) int {
		if b==0 {
			return a
		}
		return gcd(b,a%b)
	}
	func main(){
		assert(gcd(10,15),5)
		assert(gcd(15,10),5)
		assert(gcd(10,20),10)
		assert(gcd(20,10),10)
		assert(gcd(10,30),10)
		assert(gcd(30,10),10)
		assert(gcd(10,40),10)
		assert(gcd(40,10),10)
	}
	`
	ExecExpect(source)
}

func TestQuickSort(t *testing.T) {
	source := `
	func quickSort(arr []int, l int, r int) {
		if l>=r {
			return
		}
		let i=l
		let j=r
		let pivot=arr[(l+r)/2]
		while i<=j {
			while arr[i]<pivot {
				i+=1
			}
			while arr[j]>pivot {
				j-=1
			}
			if i<=j {
				let tmp=arr[i]
				arr[i]=arr[j]
				arr[j]=tmp
				i+=1
				j-=1
			}
		}
		quickSort(arr,l,j)
		quickSort(arr,i,r)
	}
	func main(){
		let arr = [5,4,3,2,1]
		quickSort(arr,0,4)
		assert(arr[0],1)
		assert(arr[1],2)
		assert(arr[2],3)
		assert(arr[3],4)
		assert(arr[4],5)
		let arr1 = [5,2,6,8,0,-2,-5,8,-7,-7]
		quickSort(arr1,0,9)
		assert(arr1[0],-7)
		assert(arr1[1],-7)
		assert(arr1[2],-5)
		assert(arr1[3],-2)
		assert(arr1[4],0)
		assert(arr1[5],2)
		assert(arr1[6],5)
		assert(arr1[7],6)
		assert(arr1[8],8)
		assert(arr1[9],8)
		let arr2 = [1,1,1,1]
		quickSort(arr2,0,3)
		assert(arr2[0],1)
		assert(arr2[1],1)
		assert(arr2[2],1)
		assert(arr2[3],1)
	}
	`
	ExecExpect(source)
}

func TestBubbleSort(t *testing.T) {
	source := `
	func bubbleSort(arr []int, len int) {
		let n=len
		for i=0;i<n-1;i+=1{
			for j=0;j<n-i-1;j+=1{
				if arr[j]>arr[j+1] {
					let tmp=arr[j]
					arr[j]=arr[j+1]
					arr[j+1]=tmp
				}
			}
		}
	}
	func main(){
		let arr = [5,4,3,2,1]
		bubbleSort(arr,5)
		assert(arr[0],1)
		assert(arr[1],2)
		assert(arr[2],3)
		assert(arr[3],4)
		assert(arr[4],5)
		let arr1 = [5,2,6,8,0,-2,-5,8,-7,-7]
		bubbleSort(arr1,10)
		assert(arr1[0],-7)
		assert(arr1[1],-7)
		assert(arr1[2],-5)
		assert(arr1[3],-2)
		assert(arr1[4],0)
		assert(arr1[5],2)
		assert(arr1[6],5)
		assert(arr1[7],6)
		assert(arr1[8],8)
		assert(arr1[9],8)
		let arr2 = [1,1,1,1]
		bubbleSort(arr2,4)
		assert(arr2[0],1)
		assert(arr2[1],1)
		assert(arr2[2],1)
		assert(arr2[3],1)
	}
	`
	ExecExpect(source)
}

func TestNarcissisticNumber(t *testing.T) {
	source := `
	func isNarcissisticNumber(n int) bool {
		let sum=0
		let m=n
		while m>0 {
			let digit=m%10
			sum+=digit*digit*digit
			m= m /10
		}
		return sum==n
	}
	func main(){
		assert_bool(isNarcissisticNumber(153),true)
		assert_bool(isNarcissisticNumber(370),true)
		assert_bool(isNarcissisticNumber(371),true)
		assert_bool(isNarcissisticNumber(407),true)
	}
	`
	ExecExpect(source)
}

func TestBinarySearch(t *testing.T) {
	source := `
	func binarySearch(arr []int, len int, x int) int {
		let l=0
		let r=len-1
		while l<=r {
			let m=l+(r-l)/2
			if arr[m]==x {
				return m
			}
			if arr[m]<x {
				l=m+1
			} else {
				r=m-1
			}
		}
		return -1
	}
	func main(){
		let arr = [2,3,4,10,40]
		let n=5
		let x=10
		let result = binarySearch(arr,n,x)
		assert(result,3)
		let x1=5
		let result1 = binarySearch(arr,n,x1)
		assert(result1,-1)
	}
	`
	ExecExpect(source)
}

func TestPalindrome(t *testing.T) {
	source := `
	func isPalindrome(s string, len int) bool {
		let l=0
		let r=len-1
		while l<r {
			if s[l]!=s[r] {
				return false
			}
			l+=1
			r-=1
		}
		return true
	}
	func main(){
		assert_bool(isPalindrome("hello",5),false)
		assert_bool(isPalindrome("helleh",6),true)
		assert_bool(isPalindrome("helleh",5),false)
	}
	`
	ExecExpect(source)
}
