// Create a Vue application
// const app = Vue.createApp({})


// Define a new global component called button-counter
Vue.component('pantry', {
    data() {

        return {
            pantry: []
        }
    },
    created() {
        fetch('/getPantry')
            .then(response => response.json())
            .then(json => {
                this.pantry = json.pantry
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