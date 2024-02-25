
func sqrtNewton(x double, epsilon double)  {
    guess = x / 2.0
    for ((guess * guess - x) >= epsilon) || ((guess * guess - x) <= -epsilon) {
        guess = guess - (guess*guess-x)/(2*guess)
    }
    return guess
}

class Edge {
	src Block
	dst Block
}

class Block {
	x int
	Succs []Block
	Preds []Block
	static id int

	func Block() {
		this.x = this.id 
		this.id += 1
	}
}

class GraphBuilder {
	blocks []Block
	edges []Edge

	func createBlock() {
		let b = new Block
		return b 
	}

	pub func buildGraph() {
		// ......
	}
}