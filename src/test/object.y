func sqrtNewton(x double, epsilon double) double {
    let guess = x / 2.0
    while ((guess * guess - x) >= epsilon) || ((guess * guess - x) <= -epsilon) {
        guess = guess - (guess*guess-x)/(2.0*guess)
    }
    return guess
}

func main(){
	cprint_double(sqrtNewton(2.0, 0.0000001))
}