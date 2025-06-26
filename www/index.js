
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

        let last_alive;
        if (target['LastUp'] != null) {
            last_alive = new Date(target['LastUp']);
        } else {
            last_alive = "Never";
        }
        row.appendChild(construct_td(last_alive));

        let last_down;
        if (target['LastDown'] != null) {
            last_down = new Date(target['LastDown']);
        } else {
            last_down = "Never";
        }
        row.appendChild(construct_td(last_down));

        let times_down = target['TimesDown'];
        row.appendChild(construct_td(times_down.length));

        let btn_td = document.createElement("td");
        let hist_btn = document.createElement("button");
        hist_btn.setAttribute("class", "btn btn-info");
        hist_btn.appendChild(document.createTextNode("🔍"));
        hist_btn.setAttribute
        hist_btn.addEventListener("click", () => {
            document.getElementById("modal-title").innerHTML = `Down Events for "${target['Name']}"`;

            const list = document.createElement("ul");
            list.setAttribute("class", "list-group");

            times_down.forEach((e) => {
                const item = document.createElement("li");
                item.setAttribute("class", "list-group-item");
                item.appendChild(document.createTextNode(new Date(e)));
                list.appendChild(item);
            });

            let modal_body = document.getElementById("modal-body");
            modal_body.innerHTML = "";
            modal_body.appendChild(list);


            modal.show();
        });
        btn_td.appendChild(hist_btn);
        row.appendChild(btn_td);

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

let modal;

window.onload = () => {

    modal = new bootstrap.Modal(document.getElementById('myModal'), "")

    get_info();

    let timer = setInterval(get_info, 10000);
} 