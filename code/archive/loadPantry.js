// Retrieve pantry items for user via GRPC request and display with Vue.js component
// written by: Mark Stanik
// tested by: Maxwell Legrand
// debugged by: Elysia Heah

const { Token } = require('./token_pb');
const { ServerClient } = require('./token_grpc_web_pb');
import Cookies from 'js-cookie'

Vue.component('pantry', {
    data() {

        return {
            pantry: [],
            lowItems: [],
            expiredItems: [],
            foodName: ""
        }
    },
    // Get pantry items and display them to the user along with any potential warnings regarding stock and expiration
    created() {
        this.foodName = foodName
        let self = this;
        let url = window.location.origin
        var service = new ServerClient(url);
        var request = new Token();
        console.log(Cookies.get("token"))
        request.setToken(Cookies.get("token"));
        console.log(request)
        service.getPantry(request, {}, function (err, response) {
            console.log("Got Response...")
            console.log(response)
            console.log(err)
            console.log(response.toObject())
            self.pantry = response.toObject().pantryList
            // Corresponds to the THHandler
            self.pantry.forEach(element => {
                let qThresh = element.quantitythreshold
                let quantity = 0.0
                var today = new Date();
                var dd = today.getDate();
                var mm = today.getMonth() + 1;
                var yyyy = today.getFullYear();
                if (dd < 10) {
                    dd = '0' + dd;
                }

                if (mm < 10) {
                    mm = '0' + mm;
                }
                today = mm + '/' + dd + '/' + yyyy;
                if (element.expiration < today) {
                    self.expiredItems.push(element)
                }
                if (element.quantity == "N/A") {
                    quantity = qThresh
                } else if (isNaN(parseFloat(element.quantity)) == false) {
                    quantity = parseFloat(element.quantity)
                }

                if (quantity < qThresh && qThresh >= 0) {
                    self.lowItems.push(element)
                }
            });
        });
    },
    template: `
    <div>
        <br>
        <div class="row">
            <div class="col-1"></div>
            <div class="col-10">
                <h1>My Digital Pantry</h1>
                <table class="table" id="myTable">
                    <tr>
                        <th></th>
                        <th onclick="sortTable(1)">Name</th>
                        <th onclick="sortTable(2)">Quantity</th>
                        <th onclick="sortTable(3)">Weight</th>
                        <th onclick="sortTable(4)">Volume</th>
                        <th onclick="sortTable(5)">Expiration</th>
                        <th></th>
                        <th></th>
                    </tr>
                    <tr v-for="pantryItem in pantry">
                        <td>
                            <template v-if='pantryItem.imagelink != ""'>
                                <img style="width: 40px; height: 40px" v-bind:src="pantryItem.imagelink">
                            </template>
                        </td>
                        <td>{{pantryItem.name}}</td>
                        <td>{{pantryItem.quantity}}</td>
                        <td>{{pantryItem.weight}}</td>
                        <td>{{pantryItem.volume}}</td>
                        <td>{{pantryItem.expiration}}</td>
                        <td><a v-bind:href="'/edit/'+ pantryItem.id">Edit</a></td>
                        <td><a v-bind:href="'/delete/'+ pantryItem.id">Delete</a></td>
                </tr>
                </table>
                <div v-if="foodName != ''" class="alert alert-info" role="alert">
                    {{foodName}}
                </div>
                <div v-for="item in expiredItems">
                    <div v-if="item.name[item.name.length -1] == 's'" class="alert alert-danger" role="alert">
                        {{item.name}} are expired
                    </div>
                    <div v-else class="alert alert-danger" role="alert">
                        {{item.name}} is expired
                    </div>
                </div>
                <div v-for="item in lowItems">
                    <div v-if="item.name[item.name.length -1] == 's'" class="alert alert-warning" role="alert">
                        {{item.name}} are in low stock
                    </div>
                    <div v-else class="alert alert-warning" role="alert">
                        {{item.name}} is in low stock
                    </div>
                </div>
                <br>
                <a href="/additem" class="btn btn-primary" role="button">Add +</a>
                <a href="/recipe" class="btn btn-primary" role="button">Find Recipes</a>
                <a href="/" class="btn btn-primary" role="button">Back</a>
            </div>
        </div>
    </div>
    `
})

new Vue({ el: '#pantryDiv' })