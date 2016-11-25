package main
  
	import (
		"encoding/json"
		"net/http"
		"github.com/julienschmidt/httprouter"
		"fmt"
		"github.com/ttacon/chalk"
		"gopkg.in/mgo.v2"
		"gopkg.in/mgo.v2/bson"
		"log"
		"html/template"
		"time"
		"strconv"
		"os"
	)


	type Idea struct{
		ID     bson.ObjectId `bson:"_id,omitempty" json:"-"`
		Body string	`json:"body" bson:"body"`
		Author string `json:"author" bson:"author"`
		Upvotes int  `json:"upvotes" bson:"upvotes"`
		Iplist []string	`json:"iplist" bson:"iplist"`
		Postip string `json:"postip" bson:"postip"`
		TimeAdded string `json:"timeadded" bson:"timeadded"`
	}

	type BigOb struct{

		IdeaObj []Idea `json:"ideaobj" bson:"ideaobj"`
		Today string `json:"today" bson:"today"`

	}

	func ServeHTMl(w http.ResponseWriter, r *http.Request, _ httprouter.Params){


			ip := r.RemoteAddr

			filepath:= r.URL.Path

			filename:=filepath[1:]

			if len(filename)==0 {
				Home(w,r,nil)
				fmt.Println(chalk.Yellow,ip," requested home page...",chalk.Reset)
			} else{  
				http.ServeFile(w,r,filename+".html")

				fmt.Println(chalk.Yellow,ip," requested",filename,"page...",chalk.Reset)	
			}
		
	}

	func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params){

       	session, err := mgo.Dial("mongodb://localhost:27017")		
       			riperr(err)
       			var all []Idea	
       			 defer session.Close()

       			 c := session.DB("ideas").C("idealist")

       			 riperr(err)

        			err = c.Find(nil).Sort("-_id").All(&all)
        			riperr(err)

					t,_ := template.ParseFiles("home.html")
		
					year,_,day := time.Now().Date()

					month := time.Now().Month().String()


					today:= strconv.Itoa(day)+" "+month+" "+strconv.Itoa(year)
					
					bigob1:= BigOb{
					IdeaObj : all,
					Today : today,
					}
					t.Execute(w,bigob1)
					fmt.Println(time.Now())

	}

	func GoBack(w http.ResponseWriter, r *http.Request, p httprouter.Params){

		http.Redirect(w, r, "/", 301)
	}

	func Submit(w http.ResponseWriter, r *http.Request, p httprouter.Params){


			author := r.FormValue("author")
			body := r.FormValue("idea")

			year,_,day := time.Now().Date()

			month := time.Now().Month().String()


			t := strconv.Itoa(day)+" "+month+" "+strconv.Itoa(year) 

		i1 := &Idea{
			Body : body,
			Author : author, 
			Upvotes :0,
			Iplist :[]string{},
			TimeAdded : t,
			}		

		output:= "Your idea is recorded "+i1.Author+i1.Postip

		fmt.Println(output)

		

         session, err := mgo.Dial("mongodb://localhost:27017")
        riperr(err)

        defer session.Close()

        c := session.DB("ideas").C("idealist")

        err = c.Insert(i1)

        riperr(err)

        http.Redirect(w,r,"/",301)

		
	}

	func Upvote(w http.ResponseWriter, r *http.Request, p httprouter.Params){

		ip := r.RemoteAddr

		
		lookFor:= Idea{}

		id:= p.ByName("id")


		obid := bson.ObjectIdHex(id)

		FindWith := bson.M{"_id":obid}	

         session, err := mgo.Dial("mongodb://localhost:27017")
        riperr(err)

        defer session.Close()

        c := session.DB("ideas").C("idealist")

        err = c.Find(FindWith).One(&lookFor)

        i:=0

        for range lookFor.Iplist{

        	if lookFor.Iplist[i] == ip[:len(string(ip))-5] {

        		i= -1
        		break
        	}
        	i++
        }

        if i!=-1{
        val,_ := json.Marshal(lookFor)

        fmt.Println(string(val))

        k := lookFor.Upvotes;

        k++

        change1 := bson.M{"$set":bson.M{"upvotes":k}}
        change2 := bson.M{"$push":bson.M{"iplist":ip[:len(ip)-5]}}
        err = c.UpdateId(obid,change1)
        riperr(err)
        err = c.UpdateId(obid,change2)

        riperr(err)
        http.Redirect(w,r,"/",http.StatusSeeOther)


    	}
    	if i==-1 {
    		fmt.Println("Only once")
    		http.Redirect(w,r,"/",http.StatusSeeOther)
    	}


	}


	func main(){
		
		newfeed:= chalk.Red.NewStyle().WithBackground(chalk.Black)

		Server:= httprouter.New()

		

		Server.POST("/submit",Submit)

		Server.GET("/submit",GoBack)

        Server.GET("/the-creator",ServeHTMl)
        Server.GET("/",ServeHTMl)
        
        Server.GET("/upvote/:id",Upvote)

        Server.ServeFiles("/resources/*filepath", http.Dir("resources"))

		fmt.Println(newfeed,"waiting at :4747",chalk.Reset);

		http.ListenAndServe(GetPort(),Server)

	}


func GetPort() string {
        var port = os.Getenv("PORT")
        // Set a default port if there is nothing in the environment
        if port == "" {
                port = "4747"
                fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
        }
        return ":" + port
}


 func riperr(err error){
 	if err!= nil{

    		log.Fatal(err)

    	}
 }
