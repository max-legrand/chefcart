// Determine if user is logged in via GRPC request & display dynamic landing page with Vue.js

const { Token } = require('./token_pb');
const { ServerClient } = require('./token_grpc_web_pb');
import Cookies from 'js-cookie'


// Define a new global component called button-counter
Vue.component('welcome-comp', {
    data() {
        return {
            username: null,
            isMobile: false,
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
        service.authUser(request, {}, function (err, response) {
            if (response == null) {
                self.username = ""
            } else {
                let data = response.toObject()
                console.log(data)
                self.username = data.token
            }
            var isMobile = false; //initiate as false
            // device detection
            if (/(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|ipad|iris|kindle|Android|Silk|lge |maemo|midp|mmp|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows (ce|phone)|xda|xiino/i.test(navigator.userAgent)
                || /1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(navigator.userAgent.substr(0, 4))) {
                isMobile = true;
            }
            self.isMobile = isMobile
            if (self.isMobile && self.username != "") {
                self.username = self.username.substring(0, self.username.indexOf("@")) + "\n" + self.username.substring(self.username.indexOf("@"))
            }
        });
    },
    template: `
    <div class="container-fluid">
            <div v-if="isMobile" class="row">
                <div v-if="username ==  '' && username != null" class="col-12 text-center">
                    <img style="max-width:100%; max-height:100%;" src="js/large_chefcart.png" />
                    <br>
                    <h1>Welcome to ChefCart</h1>
                    <h3>ChefCart is a free, open-source platform for home cooks of any level. ChefCart's goal is to provide an easy method for managing inventory, discovering new recipes, and finding ingredients you may need.</h3>
                    <br>
                </div>
                <div v-else-if="username != null" class="col-12 text-center">
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
                    <a style="width: 160px" href="/grocery" class="btn btn-primary" role="button">Grocery List</a>
                    <br>
                    <br>
                </div>
            </div>

            <div v-else class="row">
                <div class="col-sm-1"></div>
                <div v-if="username == '' && username != null" class="col-sm-10 text-center">
                    <img style="max-width:100%; max-height:100%;" src="js/large_chefcart.png" />
                    <br>
                    <h1>Welcome to ChefCart</h1>
                    <h3>ChefCart is a free, open-source platform for home cooks of any level. ChefCart's goal is to provide an easy method for managing inventory, discovering new recipes, and finding ingredients you may need.</h3>
                    <br>
                </div>
                <div  v-else-if="username != null" class="col-sm-10 text-center">
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
                    <a style="width: 160px" href="/grocery" class="btn btn-primary" role="button">Grocery List</a>
                </div>
            </div>
        </div>
    `
})

// app.mount('#welcomeDiv')
new Vue({ el: '#welcomeDiv' })