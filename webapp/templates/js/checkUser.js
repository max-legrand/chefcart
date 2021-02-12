
// 'use strict';

const e = React.createElement;

class Content extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            user: ""
        };
        this.serverRequest = this.serverRequest.bind(this);
        this.logout = this.logout.bind(this);
    }

    logout() {
        Cookies.remove("token")
        this.setState({ user: "" })
    }

    componentDidMount() {
        this.serverRequest();
    }
    serverRequest() {
        let self = this
        fetch("http://localhost:8080/authuser")
            // fetch("http://localhost:8080/authuser")
            .then(res => res.arrayBuffer())
            .then(
                (result) => {
                    protobuf.load("/js/token.proto", function (err, root) {
                        if (err)
                            throw err;
                        // Obtain a message type
                        var Token = root.lookupType("main.Token");
                        console.log("RESULT RECIEVED")
                        var uint8View = new Uint8Array(result);
                        console.log(uint8View)
                        var message = Token.decode(uint8View);
                        if (message.token == "") {
                            Cookies.remove("token")
                        }
                        console.log(message)
                        console.log(message.token)
                        self.setState({
                            user: message.token
                        });
                    });

                },
                (error) => {
                    Cookies.remove("token")
                    backtohome()
                }
            )
    }
    render() {
        console.log(this)
        if (this.state.user != "") {
            return (
                <div>
                    <img src="/js/large_chefcart.png" />
                    <br />
                    <br />
                    <h1>Logged in as: {this.state.user}</h1>
                    <br />
                    <a onClick={this.logout}><button className="btn btn-primary">Logout</button></a>
                </div>
            )

        }
        else {
            return (
                <div>
                    <img src="/js/large_chefcart.png" />
                    <br />
                    <br />
                    <a href="/signup"><button className="btn btn-primary">Signup</button></a>
                    <br />
                    <br />
                    <a href="/login"><button className="btn btn-primary">Login</button></a>
                </div>
            )
        }

    }
}

const domContainer = document.querySelector('#contentblock');
ReactDOM.render(e(Content), domContainer);