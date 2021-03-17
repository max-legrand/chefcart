// Create a Vue application
// const app = Vue.createApp({})

// Define a new global component called button-counter
Vue.component('welcome-comp', {
    data() {

        return {
            username: "",
            isMobile: false
        }
    },
    created() {
        fetch('/authuser')
            .then(response => response.json())
            .then(json => {
                if (json.token == "") {
                    this.usernam = ""
                } else {
                    this.username = json.token.Claims.name
                }
            })
            .then(() =>
                fetch('/isMobile')
                    .then(response => response.json())
                    .then(json => {
                        this.isMobile = json.isMobile
                        if (this.isMobile && this.username != "") {
                            this.username = this.username.substring(0, this.username.indexOf("@")) + "\n" + this.username.substring(this.username.indexOf("@"))
                        }
                    })
            )

    },
    template: `
    <div class="container-fluid">
            <div v-if="isMobile" class="row">
                <div v-if="username == ''" class="col-12 text-center">
                    <img style="max-width:100%; max-height:100%;" src="js/large_chefcart.png" />
                    <br>
                    <br>
                    <a style="width: 150px" class="btn btn-primary" role="button" href="/signup">Signup</a>
                    <a style="width: 150px" class="btn btn-primary" role="button" href="/login">Login</a>
                </div>
                <div v-else class="col-12 text-center">
                    <img style="max-width:100%; margin-bottom: -25px" src="js/large_chefcart.png" />
                    <h1 style="word-wrap: break-word;">Logged in as: {{ username }}</h1>
                    <br>
                    <a style="width: 160px" href="/pantry" class="btn btn-primary" role="button">Digital Pantry</a>
                    <br>
                    <br>
                    <a style="width: 160px" href="/useredit" class="btn btn-primary" role="button">Edit Account Info</a>
                    <br>
                    <br>
                    <a style="width: 160px" href="/recipe" class="btn btn-primary" role="button">Find Recipes</a>
                    <br>
                    <br>
                    <a style="width: 160px" href="/logout" class="btn btn-primary" role="button">Logout</a>
                </div>
            </div>

            <div v-else class="row">
                <div v-if="username == ''" class="col-sm-12 text-center">
                    <img src="js/large_chefcart.png" />
                    <br>
                    <br>
                    <a style="width: 150px" class="btn btn-primary" role="button" href="/signup">Signup</a>
                    <a style="width: 150px" class="btn btn-primary" role="button" href="/login">Login</a>
                </div>
                <div v-else class="col-sm-12 text-center">
                    <img src="js/large_chefcart.png" />
                    <br>
                    <h1>Logged in as:</h1>
                    <h1 class="text-wrap">{{ username }}</h1>
                    <br>
                    <a style="width: 160px" href="/pantry" class="btn btn-primary" role="button">Digital Pantry</a>
                    <a style="width: 160px" href="/useredit" class="btn btn-primary" role="button">Edit Account Info</a>
                    <br>
                    <br>
                    <a style="width: 160px" href="/recipe" class="btn btn-primary" role="button">Find Recipes</a>
                    <a style="width: 160px" href="/logout" class="btn btn-primary" role="button">Logout</a>
                </div>
            </div>
        </div>
    `
})

// app.mount('#welcomeDiv')
new Vue({ el: '#welcomeDiv' })