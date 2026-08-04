
we have update the data what i notice and i hope we change. 
1. fasttp.T // should be the fastttp.Test 
2. we change fasttp.New() for the chained or something 
3. we do the NewFunc()  
4. i want the test to mirror both types but heres the wrong thing  i see.  

### Header
1. we use HeaderMap{} here but fasttp.T has Header that 
2. Header sould have methods, Header(key, value)  we can just use Header(key, value) to set or set a value but if we want it all  we use SetHeader(....values any)  where we use by par odd|even as key, value this is easier also we can set values as one HeaderMap  

1. why? 
    a. we can just use  .Header("content-type", "application/json")
    b. we can set multiple values at the time 

### body 
body has or should have methods that  are what we use .JSON() .Data() //the raw data regardless of anything 
this is the Body{}  this is meand to be we then have request 

### Request 
is like this Response{}  like above but its for the request  we can have method, header, body,   status is empty 

why? we can check  the data we can say 
req := New().Request()

req.Header("content-type", "application/json")
we can say 
req.JSON() // internlly req.Body.JSON()   we also have Type() in the body 


this is what i want 

struct Body {
data any // this is data this is the raw data 
dataType string // "json", "xml" or otheres

Data() []b { } // this is the data as bytes 
Type() string // the dataType 
}

type Header map[string]string

Header.Header(values ...string) // this is what is it by default this is what is it  you set values its always overrides this is by pair 

Header.Set(HeaderMap{}) // this sets the whole header but yeah you can just say fasttp.T.HeaderMap = HeaderMap{}

then we have two 

type R {
Method
Header
Body 
...others
} // request 

type  T {
Method 
Header 
Body 
..others
} // test



also we can just ad others like status type 

in @status/ 
we can use  status.Ok status.BadRequest status.Created we can use ths 


also some other default methos example Content() //l for content type 

ContentJson() // this might be json 

NOTE: we dont do this yet the other methods but i want this reflected like always follow the @promt.md isntructions then  create a todo execute also note execution must be done by other agent  if you have a sequencial you just tell it you the main  agent should just be orchistrator  even better if we commit it then use worktree after finish we do question and anser just in case  to check it 

