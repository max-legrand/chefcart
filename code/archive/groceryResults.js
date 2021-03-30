// Create a Vue application
// const app = Vue.createApp({})
const { SearchQuery } = require('./token_pb');
const { ServerClient } = require('./token_grpc_web_pb');
import Cookies from 'js-cookie'

console.log("loaded")
console.log(id)

// Define a new global component called button-counter
Vue.component('groceryresults', {
    data() {

        return {
            address: null,
            monday: null,
            tuesday: null,
            wednesday: null,
            thursday: null,
            friday: null,
            saturday: null,
            sunday: null,
            distance: null,
            results: []
        }
    },
    created() {
        let self = this;
        let url = window.location.origin
        var service = new ServerClient(url);
        var request = new SearchQuery();
        console.log(Cookies.get("token"))
        request.setToken(Cookies.get("token"));
        request.setId(id);
        console.log(request)
        service.getSearchResults(request, {}, function (err, response) {
            console.log("Got Response...")
            console.log(response)
            console.log(err)
            if (err != null && err.message == "No Stores Found") {
                self.address = ""
            } else if (err != null && err.message == "Item does not belong to you") {
                window.location.href = location.origin + "/grocery"
            } else {
                let data = response.toObject()
                console.log(data)
                self.address = data.address
                self.distance = data.distance
                self.monday = data.monday
                self.tuesday = data.tuesday
                self.wednesday = data.wednesday
                self.thursday = data.thursday
                self.friday = data.friday
                self.saturday = data.saturday
                self.sunday = data.sunday
                self.results = data.resultsList
            }
            console.log(self.address)
        });
    },
    template: `
    <div>
        <br>
        <div class="row">
            <div class="col-md-2"></div>
            <div class="col-md-8">
                <h1>Search Results</h1>
                <h1 v-if="address == '' && address != null">No Stores Found</h1>
                <div v-else-if="address != null">
                    <div class="alert alert-dark" role="alert">
                        My Store: {{address}}
                        <br>
                        {{distance}} miles away
                        <br>
                        Hours: 
                        <br>
                        Mon {{monday}} | Tue {{tuesday}} | Wed {{wednesday}} | Thu {{thursday}} | Fri {{friday}} | Sat {{saturday}} | Sun {{sunday}}
                    </div>
                    <table class="table" id="myTable">
                        <tr>
                            <th></th>
                            <th onclick="sortTable(1)">Name</th>
                            <th onclick="sortTable(2)">Availability</th>
                            <th onclick="sortTable(3)">Price</th>
                            <th onclick="sortTable(4)">Rating</th>
                            <th onclick="sortTable(5)"># Reviews</th>
                            <th></th>
                        </tr>
                        <tr v-for="item in results">
                            <td>
                                <template v-if='item.image != ""'>
                                    <img style="width: 150px; height: 150px" v-bind:src="item.image">
                                </template>
                            </td>
                            <td>
                                <a v-bind:href="item.link">{{item.name}}</a>
                            </td>
                            <td v-if="item.instock == true" style="color: #57cf91">
                                In Stock
                            </td>
                            <td v-else style="color: #f56042">
                                Out of Stock
                            </td>
                            <td>
                                {{item.price.toFixed(2)}}$
                            </td>
                            <td>
                                {{item.rating.toFixed(2)}} <i class="fas fa-star"></i>
                            </td>
                            <td>
                                {{item.reviews}} reviews
                            </td>
                        </tr>
                    </table>
                </div>
                <br>
                <a href="/grocery" class="btn btn-primary" role="button">Back</a>
                <br>
                <br>
            </div>
        </div>
    </div>
    `
})

// app.mount('#pantryDiv')
new Vue({ el: '#pantryDiv' })