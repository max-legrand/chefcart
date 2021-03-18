// Create a Vue application
// const app = Vue.createApp({})

// Define a new global component called button-counter
Vue.component('recipeobj', {
    data() {

        return {
            recipes: []
        }
    },
    created() {
        let baseurl = "https://spoonacular.com/"
        dataarray = JSON.parse(dataarray)
        dataarray.forEach(element => {
            element.OriginalLink = baseurl + element.Name.replace(" ", "-") + "-" + element.ID
            this.recipes.push(element)
        });
        console.log(this.recipes.length)

    },
    template:
        `
        <div>
            <div class="row">
                <div class="col-md-12 text-center">
                    <br>
                    <h1>Recipes</h1>
                </div>
            </div>
            <div v-if="recipes.length != 0" class="row">
                <div class="col-md-2"></div>
                <div class="col-md-8">
                    <table class="table" id="myTable">
                        <tr>
                            <th></th>
                            <th onclick="sortTable(1)">Name</th>
                            <th onclick="sortTable(2)">Used Ingredients</th>
                            <th onclick="sortTable(3)">Missing Ingredients</th>
                            <th></th>
                        </tr>
                        <tr v-for="recipe in recipes">
                                <template v-if='recipe.ImageLink != ""'>
                            <td>
                                <img style="width: 40px; height: 40px" v-bind:src="recipe.ImageLink">
                            </td>
                            </template>
                            <td>
                                {{recipe.Name}}
                            </td>
                            <td>
                                {{recipe.Used}}
                            </td>
                            <td>
                                {{recipe.Missing}}
                            </td>
                            <td>
                                <a v-bind:href="recipe.OriginalLink">View Recipe</a>
                            </td>
                        </tr>
                    </table>
                    <a class="btn btn-primary" href="/recipe" role="button">Back</a>
                </div>
            </div>
            <div v-else class="row">
                <div class="col-md-2"></div>
                <div class="col-md-8">
                    <h2>No recipes found</h2>
                    <a class="btn btn-primary" href="/recipe" role="button">Back</a>
                </div>
            </div>
            <br>
            <br>
        </div>
        `
})

// app.mount('#pantryDiv')
let app = new Vue({ el: '#recipeDiv' })
