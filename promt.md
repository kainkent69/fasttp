using gin httptest we will create an abstraction to create a test for things this is where when we want to make http tests not a chore by using map based  value and optional testing 

```go
// we create a test using this 
// newTest(name,path, method, data);
// then we do this
const test = NewTest("test name", "push", data) // we use bytes
test.send(); // sends
test.header("accept"],"applcation/json")
test.status(200)
test.response.json() // know its json and returns a test.data 
test.data(func(data){
data.ok = true;
data.data  = Data{}
// or if you are kind of  lazy 
data = NewData() // this is what the expected is
})

// but we can do it for more abstracted 
NewFuncTest("test name", "/testpath", nil, func(test) TestInfo {
return TestInfo{
	Status: 200,
	Json: true // this ceate a data internal data but can be accessed
	headers: HeaderMap{
		"content-type":"application/json",
		"x": "this is x"
	}
	data : func(data) {
		data.sucess = false, 
		data.data = NewData()
	}
}
})

```

internally we use require and assert just to make life easy 

### Test

```go
// we create methods in ./tests this are the resuablethings that we want to call we can say 
/// things like test.Start() 

func Check(t *testing.T){
// this is an abstactions this means to be resuable especially if its complex andmultipe
// this is abstractions so it is allowed to have condition to do things in one place
}
#// we group it with this
func TestName(t *testing.T){
// errors or without errors should be checked by testlify require
require.NoError(t, err);
// if we have to assert we dont do it individually we either create a standalone types 
o// or we assert them directly 
aassert.Equal(t, expectedStruct, actaulStruct) // this makes code cleaner
w// we should create functions and abstractions in tests espeically if its cratble this also 
///makes changing tests better 
}
// if we have sub edge to the spicific scope we do 
t.Run("test", func(t *testing.T){
// the tests
})
}

```

I want yo to do this. 
have the main agent as orchistrator and control let this able to do things like checking or  wieging and repromting based on the code, to verify run a small tests or create a sample  project in @sample this 

the execution i think it is good to have the agent does a git worktree of its own, based on the informations given. 

ex: find the problem in a check if 1. can do x, 2. can do y, tset pass, say do it runs the test you said it would  then it notify you the main then checkout hten run the test, then checkout then run the test, or pull this  is a way for us to develop the project in the best way possible  

so main is empty, it now then commit, hten launches a worktree or launches a workflow in which  you tend to check also test should be targeted. 

this way you have agents that does its thing  then after it finishes you check thier finding, also while building have each add or update changeme.md and also update changelog or  a backtrack.md its all the  or have a readme[branch or worktree].md this you can append but its best to have an append only file then based in append only file you  create the  readme.md also you can launch paralelally or you can launch them sequencially, but i think if it is sequencial no need to create a separate worktree just do a list of todo then work with it read this do nothing and wait for my other instruction to continue
