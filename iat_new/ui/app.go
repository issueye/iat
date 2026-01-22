package main

// app.go is merged into main.go for simplicity in this refactor, 
// or can be kept separate. 
// Since I defined App struct in main.go, I don't strictly need app.go unless for more methods.
// I will leave this file as a placeholder or specific bridge methods.

func (a *App) Greet(name string) string {
    return "Hello " + name
}
