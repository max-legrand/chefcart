// Pull user data via GRPC request and dynamically modify recipe request form

const { Token } = require('./token_pb');
const { ServerClient } = require('./token_grpc_web_pb');
import Cookies from 'js-cookie'


// Define a new global component called button-counter
Vue.component('recipe', {
    data() {

        return {
            pantry: [],
            isMobile: false,
            selected: [],
            diets: [],
            intolerances: []
        }
    },
    methods: {
        checkInputs: function (e) {
            if (this.selected.length == 0 && $("#additionalIngredients").val() == "") {
                alert("Please add an ingredient to search for")
            } else {
                return true
            }
            e.preventDefault();
        },
        addItem: function (event) {
            if ($("#" + event.target.id).prop("checked") == true) {
                // console.log(event.target.id)
                this.selected.push(event.target.id)
            } else {
                const index = this.selected.indexOf(event.target.id);
                if (index > -1) {
                    this.selected.splice(index, 1);
                }
            }
        },
        stopInput: function (event) {
            alert("enter clicked")
            event.preventDefault()
        }
    },
    created() {
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
        });
        service.getUserInfo(request, {}, function (err, response) {
            let userinfo = response.toObject()
            console.log(userinfo)
            self.diets = userinfo.dietsList
            self.intolerances = userinfo.intolerancesList
            $('#diets').selectpicker();
            $('#intolerances').selectpicker();
            $('#diets').val(self.diets)
            $('#diets').selectpicker("refresh");
            $('#intolerances').val(self.intolerances)
            $('#intolerances').selectpicker("refresh");
            $('#additionalIngredients').tagsinput({
                cancelConfirmKeysOnEmpty: true
            });
            $('#tags-input').tagsinput({
                confirmKeys: [13, 188]
            });

            $('#tags-input input').on('keypress', function (e) {
                if (e.keyCode == 13) {
                    e.keyCode = 188;
                    e.preventDefault();
                };
            });
            $('.bootstrap-tagsinput input').keydown(function (event) {
                if (event.which == 13) {
                    $(this).blur();
                    $(this).focus();
                    return false;
                }
            })
        })

    },
    template:
        `
        <div>
            <div class="row">   
                <div class="col-md-12 text-center">
                    <br>
                    <h1>Recipe Finder</h1>
                </div>
            </div>
            <div class="row">   
                <div class="col-md-2"></div>
                <div class="col-md-8">
                    <form method="POST" v-on:submit="checkInputs" action="/recipeSearch">
                        <br>
                        <h2>Pantry Ingredients</h2>
                        <table id="myTable" class="table">
                            <tr>
                                <th></th>
                                <th onclick="sortTable(1)">Name</th>
                                <th onclick="sortTable(2)">Quantity</th>
                                <th>Add to Search</th>
                            </tr>
                            <tr v-for="pantryItem in pantry">
                                <template v-if='pantryItem.imagelink != ""'>
                                    <td>
                                        <img style="width: 40px; height: 40px" v-bind:src="pantryItem.imagelink">
                                    </td>
                                </template>
                                <td>
                                    {{pantryItem.name}}
                                </td>
                                <td>
                                    {{pantryItem.quantity}}
                                </td>
                                <td v-if="isMobile">
                                    <input v-on:click="addItem" class="form-check-input" name="ingredients[]" v-bind:id="pantryItem.name" v-bind:value="pantryItem.name" type="checkbox">
                                </td>
                                <td v-else style="padding-left: 75px">
                                    <input v-on:click="addItem" class="form-check-input" name="ingredients[]" v-bind:id="pantryItem.name" v-bind:value="pantryItem.name" type="checkbox">
                                </td>
                            </tr>
                        </table>
                        <br>
                        <h2>Additional Ingredients</h2>
                        <div class="form-group" style="width: 100%">
                            <input v-on:keyup.enter="stopInput" type="text" class="form-control" id="additionalIngredients" aria-describedby="emailHelp" name="additionalIngredients" data-role="tagsinput">
                        </div>
                        <h2>Excluded Ingredients</h2>
                        <div class="excludedIngredients form-group" style="width: 100%">
                            <input v-on:keyup.enter="stopInput" type="text" class="form-control" id="excludedIngredients" aria-describedby="emailHelp" name="excludedIngredients" data-role="tagsinput">
                        </div>
                        <h2>Specific Cuisines</h2>
                        <div class="form-group" style="width: 100%">
                            <select style="width: 100% !important" id="cuisine" name="cuisines" class="selectpicker form-control"
                                data-live-search="true" data-selected-text-format="count > 3" multiple="multiple">
                                <option val="African">African</option>
                                <option val="American">American</option>
                                <option val="British">British</option>
                                <option val="Cajun">Cajun</option>
                                <option val="Caribbean">Caribbean</option>
                                <option val="Chinese">Chinese</option>
                                <option val="Eastern European">Eastern European</option>
                                <option val="European">European</option>
                                <option val="French">French</option>
                                <option val="German">German</option>
                                <option val="Greek">Greek</option>
                                <option val="Indian">Indian</option>
                                <option val="Irish">Irish</option>
                                <option val="Italian">Italian</option>
                                <option val="Japanese">Japanese</option>
                                <option val="Jewish">Jewish</option>
                                <option val="Korean">Korean</option>
                                <option val="Latin American">Latin American</option>
                                <option val="Mediterranean">Mediterranean</option>
                                <option val="Mexican">Mexican</option>
                                <option val="Middle Eastern">Middle Eastern</option>
                                <option val="Nordic">Nordic</option>
                                <option val="Southern">Southern</option>
                                <option val="Spanish">Spanish</option>
                                <option val="Thai">Thai</option>
                                <option val="Vietnamese">Vietnamese</option>
                            </select>
                            <small class="form-text text-muted">Leave empty for all cuisines</small>
                        </div>
                        <div class="form-group" style="width: 100% !important">
                            <h2 for="restrictions">Diets</h2>
                            <br>
                            <select style="width: 100% !important" id="diets" name="diets" class="selectpicker form-control"
                                data-live-search=" true" data-selected-text-format="count > 3" multiple="multiple">
                                <option value="Gluten Free">Gluten Free</option>
                                <option value="Ketogenic">Ketogenic</option>
                                <option value="Vegetarian">Vegetarian</option>
                                <option value="Lacto-Vegetarian">Lacto-Vegetarian</option>
                                <option value="Ovo-Vegetarian">Ovo-Vegetarian</option>
                                <option value="Vegan">Vegan</option>
                                <option value="Pescetarian">Pescetarian</option>
                                <option value="Paleo">Paleo</option>
                                <option value="Primal">Primal</option>
                                <option value="Whole30">Whole30</option>
                            </select>
                        </div>
                        <div class="form-group" style="width: 100% !important">
                            <h2 for="restrictions">Intolerances</h2>
                            <br>
                            <select style="width: 100% !important" id="intolerances" name="intolerances"
                                class="selectpicker form-control" data-live-search=" true" data-selected-text-format="count > 3"
                                multiple="multiple">
                                <option value="Dairy">Dairy</option>
                                <option value="Egg">Egg</option>
                                <option value="Gluten">Gluten</option>
                                <option value="Grain">Grain</option>
                                <option value="Peanut">Peanut</option>
                                <option value="Seafood">Seafood</option>
                                <option value="Sesame">Sesame</option>
                                <option value="Shellfish">Shellfish</option>
                                <option value="Soy">Soy</option>
                                <option value="Sulfite">Sulfite</option>
                                <option value="Tree Nut">Tree Nut</option>
                                <option value="Wheat">Wheat</option>
                                </option>
                            </select>
                        </div>
                        <br>
                        <br>
                        <button type="submit" class="btn btn-primary">Submit</button>
                        <a class="btn btn-primary" href="/" role="button">Back</a>
                    </form>
                </div>
            </div>
            <br>
            <br>
        </div>
        `
})

// app.mount('#pantryDiv')
let app = new Vue({ el: '#recipeDiv' })

$(document).ready(function () {
    $('#additionalIngredients').tagsinput({
        cancelConfirmKeysOnEmpty: true
    });
    $('#excludedIngredients').tagsinput({
        cancelConfirmKeysOnEmpty: true
    });
    $('#tags-input').tagsinput({
        confirmKeys: [13, 188]
    });
    $('.bootstrap-tagsinput input').keydown(function (event) {
        if (event.which == 13) {
            $(this).blur();
            $(this).focus();
            return false;
        }
    })
    $('#tags-input input').on('keypress', function (e) {
        if (e.keyCode == 13) {
            e.keyCode = 188;
            e.preventDefault();
        };
    });
    $('#additionalIngredients').on('itemAdded', function (event) {
        $('#additionalIngredients').tagsinput('refresh');
    });
    $('#excludedIngredients').on('itemAdded', function (event) {
        $('#excludedIngredients').tagsinput('refresh');
    });
    $('#cuisine').selectpicker();
})
