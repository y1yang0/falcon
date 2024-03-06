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