package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	links := []string{
		"http://google.com",
		"http://facebook.com",
		"http://stackoverflow.com",
		"http://golang.org",
		"http://amazon.com",
	}

	c := make(chan string)

	for _, link := range links {
		go checkLink(link, c)
	}

	// for { // infinite loop
	// 	go checkLink(<-c, c)
	// }

	for l := range c { // Alternative syntax of the above. It waits for the channel to return some value to assign to l, then run the body of the for loop. Providing more clarity to other engineers
		go func(link string) {
			time.Sleep(5 * time.Second)
			checkLink(link, c)
		}(l)
	}

	// fmt.Println(<-c) // If we put one extra, our program will hang because the main routine would be sitting there waiting for someone to send some information into our channel
}

func checkLink(link string, c chan string) {
	_, err := http.Get(link) // Blocking call. When this runs, Main Go routine can do nothing else!
	if err != nil {
		fmt.Println(link, "might be down!")
		c <- link
		return
	}

	fmt.Println(link, "is up!")
	c <- link
}

// Notes:
// 1. How our code is being executed right now:
//    With our slice of links -> Take first link from slice -> Make request -> GET http://google.com -> Wait for a response, log it
//							  -> Take next link, make request -> GET http://facebook.com -> Wait for a response, log it
//							  -> Repeat for others
//
//    Basically it's in series, every single time we're making a request, We sit around and wait for the response to come back before making the next
//    So our aim to to make requests in parallel
//
// 2. When we launch a Go program, i.e. when we compile it and execute it, we automatically create one Go routine. You can think of a routine as being
//    something that exists inside of our running program or our process. This go routine takes every line of code inside of our program and executes them one by one.
//    Actual compiled form of our code might look a little bit differently than what we have.
//
// 3. Syntax of a Go routine:
//    go checkLink(link)
//    go - creates a new thread go routine
//    checkLink(link) - the function the newly created thread would run
//
// 4. What happens when we spawn multiple Go routines inside our program?
//                       One CPU Core
//                             |
//                             v
//                       Go Scheduler                      <- Scheduler runs one routine until it finishes or makes a blocking call (like an HTTP request)
//           / \              / \              / \
//            |                |                |
//            V                V                V
//       Go Routine       Go Routine       Go Routine
//
//   The most important thing to understand here is that even though we are launching multiple routines, only one is being executed or running at any given time.
//   So the purpose of this Go scheduler is to monitor the code that is running inside of each of these Go Routines. As soon as the scheduler detects that one routine
//   has finished running all of the code inside of it, so essentially all the code inside of a given function or when the scheduler it detects that a function has made a
//   blocking call like the HTTP request that we are making then it says okay you know what? You Go routine right here, you thing that just finished or has some blocking
//   code that is being executed. You're done for right now. We are going to pause you. And instead we're going to start executing this other Go routine. So essentially
//   even though we are spawning multiple Go routines, they are not actually being executed truly at the same time. Whenever we have one CPU, so this one CPU is only running
//   the code inside of one Go routine at a time and we rely upon this go scheduler to decide which Go routine is being executed. So in the blink of an eye like we might run
//   this routine right here for a fraction, then a fraction of a second and then jump over to that and then jump back over to this one. Thus, the scheduler is working very
//   quickly behind the scenes to handle all these different routines as best as it can and cycle through them very very quickly.
//
// 5. What happens when we have multiple CPU cores on our local machine?
//    By default, Go tries to only use ONE core, but we can easily change that
//    One CPU Core       One CPU Core      One CPU Core
//       |  / \             |  / \           |  / \
//       v   |              v   |            v   |
//                       Go Scheduler                      <- Scheduler runs one thread on each "logical" core
//           / \              / \              / \
//            |                |                |
//            V                V                V
//       Go Routine       Go Routine       Go Routine
//
//   when we have multiple CPU cores, each one can run one single Go routine at a time.
//   And so the Go scheduler might say oh okay we've got three separate routines and we have three separate CPU cores.
//   So rather than monitoring each routine and attempting to run only one at a time, the scheduler will instead assign one routine
//   to this core, another one to the second core and the last one to the third core. So soon as we have multiple CPU cores then
//   we're talking about running multiple chunks of code truly concurrently.
//
// 6. Concurrency - we can have multiple threads executing code. If one thread blocks, another one is picked up and worked on
//                         One Core
//                             |
//                             v
//                    Pick one Go routine
//           / \              / \              / \
//            |                |                |
//            V                V                V
//       Go Routine       Go Routine       Go Routine
//
//   So when we say something is concurrent we are simply saying that our program has the ability to run  different things kind of at the same time
//   but not really at the same time because we have one core. We're only picking one Go routine. So all we're saying with concurrency is that we can
//   kind of schedule work to be done throughout each other. We don't necessarily have to wait for one Go routine to finish before going onto the next one.
//
// 7. Parallelism = multiple threads executed at the exact same time, like nanosecond. Requires multiple CPUs
//                         One Core                                                One Core
//                             |                                                       |
//                             v                                                       v
//             Pick     one      Go     routine                        Pick     one      Go     routine
//           / \              / \              / \                   / \              / \              / \
//            |                |                |                     |                |                |
//            V                V                V                     V                V                V
//       Go Routine       Go Routine       Go Routine            Go Routine       Go Routine       Go Routine
//
// 8. Bug we're going to see as soon as we implement Go routines:
//    Our Running Program
//    Main routine - created when we launched the program
//    Child Go routine  ---\
//    Child Go routine  -------> Child routines created by 'go' keyword
//    Child Go routine  ---/
//
//    Child routines are not quite given the same level of respect, I guess for lack of a better term, we'll say respect as the main routine is.
//    If the lifespan of the main routine is shorter than our Child Go routine, the entire program quits automatically! In our case, nothing
//    got printed out...
//
//    To address that issue, we need to use a channel to make sure that the main routine is aware of when each of these child go routines have
//    completed their code. So essentially we're going to create one channel and that channel is going to communicate between all of these different
//    routines and channels are the only way that we have to communicate between go routines. There's no other way.
//
//                         Child go routine
//
//                      (bi-directional arrow)
//
//   Main Routine     <->     Channel     <->      Child go routine
//
//                      (bi-directional arrow)
//
//                         Child go routine
//
//    We only communicate using channels, so we can kind of think of a channel as being something like the above. That kind of intermediates discussion or communication between all these different
//    running routines on our local machine. You can think of the channel itself as being like text messaging or like instant messaging. So we can send some data into a channel and that will
//    automatically get sent to any other running routine on a machine that has access to that channel. We can treat a channel just like any other value inside of go. So we create a channel
//    essentially in the same way that we create a struct or a slice or an int or a string. So there are actual values that we can pass around and in this case we'll pass around to these different
//    Go routines.
//
//    Now the most important thing to understand about channels is that they are typed just like every other variable.
//    So instructor isn't just saying the fact that hey this value is of type channel. He meant to say that the information that we pass into a channel or the data that we attempt to share
//    between these different routines must all be of the same type. So essentially when we create a channel we say make a channel that is meant for sharing say type String (can be other, too)
//    throughout our application, so we will make a channel of type string.
//
//       Go Routine
//           / \
//            |   "asdf"
//            v
//  Channel of type String
//           / \
//            |   "asdf"
//            V
//       Go Routine
//
// 9. Sending Data with Channels (the syntax)
//    channel <- 5                  - Send the value '5' into this channel
//    myNumber <- channel           - Wait for a value to be sent into the channel. When we get one, assign the value to 'myNumber'
//    fmt.Println(<- channel)       - Wait for a value to be sent into the channel. When we get one, log it out immediately
//
//    This is how we send data through channels. Remember that our channel is kind of like a two way messaging device we can think of it as being like text messaging
//    So there's always going to be one person who is sending a message and then another person or another entity, i.e. our program who is receiving that message
//    For us, we might want to send data from the main routine to all of our child go routines or we might want to send data from our routine and receive it over inside of the main routine
//
// 10. Function Literals
//     JavaScript - Anonymous Function
//     Ruby - Lambda
//     Python - Lambda
//     C# - Lambda
//     PHP - Anonymous Function
//     Go - Function Literal
//
// 11. With the last commit, I could see that only http://amazon.com is up! was printing towards the end of my program...
//     Simplified Version to Illustrate Channels Gotcha
//          RAM
//     0000  | amazon.com <------  Main Routine has a variable l that is pointing at some location inside of memory that holds
//     0001  |                |
//     0002  |                L    Child Routine also pointing at the exact same memory address
//
//     So tte warning message that we've seen right there was essentially saying hey you are referencing a variable declared in the outer scope of this function literal (which is constantly changing)
//     and that's a big deal because we are trying to reference a variable that is being maintained or used by another go routine.
//     So in practice we never ever attempt to reference the same variable inside of two different routines. So we need to pass by value, we add argument to our function literal.
//     With that updated, now l can change as much as it pleases and we don't have to worry about still having our routine referencing that same copy or same address in memory.
//
//     Remember the big takeaway with routines that we just learned right here is that we should never ever try to access the same variable from a different child routine where ever possible.
//     We only share information with a child routine or a new routine that we create by passing it in as an argument or communicating with the child routine over channels.

// Quiz 11: Channels and Go Routines
// 1. Go Routines and Channels are tough, so let's start with the basics! Which of the following best describes what a go routine is? A separate line of code execution that can be used to handle blocking code
// 2. What is the purpose of a channel? For communication between go routines
// 3. Take a look at the following program. Are there any issues with it? Both answers #2 and #3 are correct:
//    2 - The greeting variable is referenced directly in the go routine, which might lead to issues if we eventually start to change the value of greeting
//    3 - The program will likely exit before the fmt.Println function has an opportunity to actually print anything out to the terminal; this might not be the intent of the program
/* 	package main

	import (
		"fmt"
	)

	func main() {
 		greeting := "Hi There!"

 		go (func() {
     		fmt.Println(greeting)
 		})()
	} */
// 4. Here's a tough one - Is there any issue with the following code? The channel is expecting values of type string, but we're passing in a value of type byte slice, which is technically not a string
/* 	package main

	func main() {
 		c := make(chan string)
 		c <- []byte("Hi there!")
	} */
// 5. Another tough one! Is there any issue with the following code? The syntax of this program is OK, but the program will never exit because it will wait for something to receive the value we're passing
//    into the channel. e.g., fatal error: all goroutines are asleep - deadlock!
/* 	package main

	func main() {
     	c := make(chan string)
     	c <- "Hi there!"
	} */
// 6. Ignoring whether or not the program will exit correctly, are the following two code snippets equivalent? They are the same
/* Snippet 1
package main

import "fmt"

func main() {
 c := make(chan string)
 for i := 0; i < 4; i++ {
     go printString("Hello there!", c)
 }

 for s := range c {
     fmt.Println(s)
 }
}

func printString(s string, c chan string) {
 fmt.Println(s)
 c <- "Done printing."
} */
/* Snippet 2
package main

import "fmt"

func main() {
 c := make(chan string)

 for i := 0; i < 4; i++ {
     go printString("Hello there!", c)
 }

 for {
     fmt.Println(<- c)
 }
}

func printString(s string, c chan string) {
 fmt.Println(s)
 c <- "Done printing."
} */
