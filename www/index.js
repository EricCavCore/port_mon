
function construct_td(content) {
    let td = document.createElement("td");
    td.appendChild(document.createTextNode(content));
    return td;
}

function construct_up_or_down(alive) {
    let td = document.createElement("td");
    let badge = document.createElement("span");
    if (alive == true) {
        badge.setAttribute("class", "badge text-bg-success")
        badge.appendChild(document.createTextNode("UP"))
    } else {
        badge.setAttribute("class", "badge text-bg-danger")
        badge.appendChild(document.createTextNode("DOWN"))
    }
    td.appendChild(badge);

    return td;
}

function build_table(data) {
    let table_body = document.querySelector("#table_body");
    table_body.innerHTML = "";

    data.forEach(target => {

        let row = document.createElement("tr");

        row.appendChild(construct_td(target['Name']));
        row.appendChild(construct_td(target['Address']));
        row.appendChild(construct_td(target['Protocol']));
        row.appendChild(construct_up_or_down(target['Alive']));

        let last_alive = new Date(target['LastAlive']);
        row.appendChild(construct_td(last_alive));
        row.appendChild(construct_td(target['TimesDown']));

        table_body.appendChild(row);
    });
    
}

function get_info() {
    let xhr = new XMLHttpRequest();
    let url = '/stats';
    xhr.open("GET", url, true);

    xhr.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            let data = JSON.parse(this.responseText);

            data.sort((a, b) => a['Name'].localeCompare(b['Name']));

            build_table(data);
            console.log("got new info");
        }
    }
    // Sending our request 
    xhr.send();
}

window.onload = () => {

    get_info();

    let timer = setInterval(get_info, 10000);
} 