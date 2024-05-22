package main

func init() {
    go StartServer()
}

func main() {
    init()
    select {} 
}
