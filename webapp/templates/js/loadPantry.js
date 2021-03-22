// Create a Vue application
// const app = Vue.createApp({})


// Define a new global component called button-counter
Vue.component('pantry', {
    data() {

        return {
            pantry: [],
            lowItems: [],
            qThresh: 0.0,
            vThresh: 0.0,
            wThresh: 0.0,
            foodName: ""
        }
    },
    created() {
        this.foodName = foodName
        fetch('/getPantry')
            .then(response => response.json())
            .then(json => {
                this.pantry = json.pantry
            })
            .then(() => {
                fetch('/getUserinfo')
                    .then(response => response.json())
                    .then(json => {
                        // console.log(json.userinfo)
                        this.qThresh = json.userinfo.QuantityThreshold
                        this.vThresh = json.userinfo.VolumeThreshold
                        this.wThresh = json.userinfo.WeightThreshold

                        this.pantry.forEach(element => {

                            let weight = 0.0
                            let volume = 0.0
                            let quantity = 0.0

                            if (element.Weight == "N/A") {
                                weight = this.wThresh
                            } else if (isNaN(parseFloat(element.Weight)) == false) {
                                weight = parseFloat(element.Weight)
                            }

                            if (element.Volume == "N/A") {
                                volume = this.wThresh
                            } else if (isNaN(parseFloat(element.Volume)) == false) {
                                volume = parseFloat(element.Volume)
                            }

                            if (element.Quantity == "N/A") {
                                quantity = this.wThresh
                            } else if (isNaN(parseFloat(element.Quantity)) == false) {
                                quantity = parseFloat(element.Quantity)
                            }

                            if ((quantity < this.qThresh && this.qThresh >= 0) || (volume < this.vThresh && this.vThresh >= 0) || (weight < this.wThresh && this.wThresh >= 0)) {
                                this.lowItems.push(element)
                            }
                        });

                    })
            })
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
                            <template v-if='pantryItem.ImageLink != ""'>
                                <img style="width: 40px; height: 40px" v-bind:src="pantryItem.ImageLink">
                            </template>
                        </td>
                        <td>{{pantryItem.Name}}</td>
                        <td>{{pantryItem.Quantity}}</td>
                        <td>{{pantryItem.Weight}}</td>
                        <td>{{pantryItem.Volume}}</td>
                        <td>{{pantryItem.Expiration}}</td>
                        <td><a v-bind:href="'/edit/'+ pantryItem.ID">Edit</a></td>
                        <td><a v-bind:href="'/delete/'+ pantryItem.ID">Delete</a></td>
                </tr>
                </table>
                <div v-if="foodName != ''" class="alert alert-info" role="alert">
                    {{foodName}} is not a vaid food item
                </div>
                <div v-for="item in lowItems">
                    <div v-if="item.Name[item.Name.length -1] == 's'" class="alert alert-warning" role="alert">
                        {{item.Name}} are in low stock
                    </div>
                    <div v-else class="alert alert-warning" role="alert">
                        {{item.Name}} is in low stock
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

// app.mount('#pantryDiv')
new Vue({ el: '#pantryDiv' })