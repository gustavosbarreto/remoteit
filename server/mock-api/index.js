const fs = require("fs");

module.exports = () => {
    return JSON.parse(fs.readFileSync("db.json"));    
}
