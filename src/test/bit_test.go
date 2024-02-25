package test

import "testing"

func TestCount1Bits(t *testing.T) {
	source := `
	func count1Bits(n int) int {
		let cnt=0
		while n>0 {
			cnt+=n%2
			n=n/2
		}
		return cnt
	}
	func main(){
		assert(count1Bits(0),0)
		assert(count1Bits(1),1)
		assert(count1Bits(2),1)
		assert(count1Bits(3),2)
		assert(count1Bits(4),1)
		assert(count1Bits(5),2)
		assert(count1Bits(6),2)
		assert(count1Bits(7),3)
		assert(count1Bits(8),1)
		assert(count1Bits(9),2)
		assert(count1Bits(10),2)
		assert(count1Bits(11),3)
		assert(count1Bits(12),2)
		assert(count1Bits(13),3)
		assert(count1Bits(14),3)
		assert(count1Bits(15),4)
	}
	`
	ExecExpect(source)
}

func TestBitAverage(t *testing.T) {
	source := `
	func average(a int, b int) int {
		return (a&b)+((a^b)>>1)
	}
	func main(){
		assert(average(1,2),1)
		assert(average(2,3),2)
		assert(average(3,4),3)
		assert(average(4,5),4)
		assert(average(5,6),5)
		assert(average(6,7),6)
		assert(average(7,8),7)
		assert(average(8,9),8)
		assert(average(9,10),9)
		assert(average(10,11),10)
		assert(average(11,12),11)
		assert(average(12,13),12)
		assert(average(13,14),13)
		assert(average(14,15),14)
	}
	`
	ExecExpect(source)
}

func TestAbsByBit(t *testing.T) {
	source := `
	func abs(n int) int {
		let mask=n>>31
		return (n+mask)^mask
	}
	func main(){
		assert(abs(0),0)
		assert(abs(1),1)
		assert(abs(-1),1)
		assert(abs(2),2)
		assert(abs(-2),2)
		assert(abs(3),3)
		assert(abs(-3),3)
		assert(abs(4),4)
		assert(abs(-4),4)
		assert(abs(5),5)
		assert(abs(-5),5)
		assert(abs(6),6)
		assert(abs(-6),6)
		assert(abs(7),7)
		assert(abs(-7),7)
	}
	`
	ExecExpect(source)
}

func TestNegateByBit(t *testing.T) {
	source := `
	func negate(n int) int {
		return ~n+1
	}
	func main(){
		assert(negate(0),0)
		assert(negate(1),-1)
		assert(negate(-1),1)
		assert(negate(2),-2)
		assert(negate(-2),2)
		assert(negate(3),-3)
		assert(negate(-3),3)
		assert(negate(4),-4)
		assert(negate(-4),4)
		assert(negate(5),-5)
		assert(negate(-5),5)
		assert(negate(6),-6)
		assert(negate(-6),6)
		assert(negate(7),-7)
		assert(negate(-7),7)
	}
	`
	ExecExpect(source)
}

func TestIsSigned(t *testing.T) {
	source := `
	func isSigned(n int) bool {
		return n>>31!=0
	}
	func main(){
		assert_bool(isSigned(0),false)
		assert_bool(isSigned(1),false)
		assert_bool(isSigned(-1),true)
		assert_bool(isSigned(2),false)
		assert_bool(isSigned(-2),true)
		assert_bool(isSigned(3),false)
		assert_bool(isSigned(-3),true)
		assert_bool(isSigned(4),false)
		assert_bool(isSigned(-4),true)
		assert_bool(isSigned(5),false)
		assert_bool(isSigned(-5),true)
		assert_bool(isSigned(6),false)
		assert_bool(isSigned(-6),true)
		assert_bool(isSigned(7),false)
		assert_bool(isSigned(-7),true)
	}
	`
	ExecExpect(source)
}
