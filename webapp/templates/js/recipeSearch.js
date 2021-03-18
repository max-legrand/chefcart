// Create a Vue application
// const app = Vue.createApp({})

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
        fetch('/getPantry')
            .then(response => response.json())
            .then(json => {
                this.pantry = json.pantry
            }).then(() =>
                fetch('/isMobile')
                    .then(response => response.json())
                    .then(json => {
                        this.isMobile = json.isMobile
                        if (this.isMobile && this.username != "") {
                            this.username = this.username.substring(0, this.username.indexOf("@")) + "\n" + this.username.substring(this.username.indexOf("@"))
                        }
                    })
            ).then(() => {
                fetch('/getDietsIntol')
                    .then(response => response.json())
                    .then(json => {
                        this.diets = json.Diets
                        this.intolerances = json.Intolerances
                        $('#diets').selectpicker();
                        $('#intolerances').selectpicker();
                        $('#diets').val(this.diets)
                        $('#diets').selectpicker("refresh");
                        $('#intolerances').val(this.intolerances)
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
                                <template v-if='pantryItem.ImageLink != ""'>
                                    <td>
                                        <img style="width: 40px; height: 40px" v-bind:src="pantryItem.ImageLink">
                                    </td>
                                </template>
                                <td>
                                    {{pantryItem.Name}}
                                </td>
                                <td>
                                    {{pantryItem.Quantity}}
                                </td>
                                <td v-if="isMobile">
                                    <input v-on:click="addItem" class="form-check-input" name="ingredients[]" v-bind:id="pantryItem.Name" v-bind:value="pantryItem.Name" type="checkbox">
                                </td>
                                <td v-else style="padding-left: 75px">
                                    <input v-on:click="addItem" class="form-check-input" name="ingredients[]" v-bind:id="pantryItem.Name" v-bind:value="pantryItem.Name" type="checkbox">
                                </td>
                            </tr>
                        </table>
                        <br>
                        <h2>Additional Ingredients</h2>
                        <div class="form-group" style="width: 100%">
                            <input v-on:keyup.enter="stopInput" type="text" class="form-control" id="additionalIngredients" aria-describedby="emailHelp" name="additionalIngredients" data-role="tagsinput">
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
    $('#cuisine').selectpicker();
})
