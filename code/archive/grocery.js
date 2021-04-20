// Display grocery list items
// written by: Shreyas Heragu
// tested by: Kevin Lin
// debugged by: Milos Seskar

const { Token } = require('./token_pb');
const { ServerClient } = require('./token_grpc_web_pb');
import Cookies from 'js-cookie'

Vue.component('grocery', {
    data() {
        return {
            pantry: [],
            foodName: ""
        }
    },
    // Get all pantry items, and foodname if an invalid item was entered previously
    created() {
        this.foodName = foodName
        let self = this;
        let url = window.location.origin
        var service = new ServerClient(url);
        var request = new Token();
        console.log(Cookies.get("token"))
        request.setToken(Cookies.get("token"));
        console.log(request)
        service.getGroceries(request, {}, function (err, response) {
            console.log("Got Response...")
            console.log(response)
            console.log(err)
            console.log(response.toObject())
            self.pantry = response.toObject().pantryList
        });
    },
    // Display content
    template: `
    <div>
        <br>
        <div class="row">
            <div class="col-md-2"></div>
            <div class="col-md-8">
                <h1>My Grocery List</h1>
                <table class="table" id="myTable">
                    <tr>
                        <th></th>
                        <th onclick="sortTable(1)">Name</th>
                        <th></th>
                        <th></th>
                    </tr>
                    <tr v-for="pantryItem in pantry">
                        <td>
                            <template v-if='pantryItem.imagelink != ""'>
                                <img style="width: 50px; height: 50px" v-bind:src="pantryItem.imagelink">
                            </template>
                        </td>
                        <td>{{pantryItem.name}}</td>
                        <td><a v-bind:href="'/search/'+ pantryItem.id">Search</a></td>
                        <td><a v-bind:href="'/deleteGrocery/'+ pantryItem.id">Delete</a></td>
                </tr>
                </table>
                <div v-if="foodName != ''" class="alert alert-info" role="alert">
                    {{foodName}}
                </div>
                <br>
                <a href="/addGrocery" class="btn btn-primary" role="button">Add +</a>
                <a href="/" class="btn btn-primary" role="button">Back</a>
            </div>
        </div>
    </div>
    `
})

new Vue({ el: '#pantryDiv' })