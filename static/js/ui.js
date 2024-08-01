function getStyleSheets() {
    let styleSheets = [];
    for (var i=0; i<document.styleSheets.length; i++) {
        styleSheets.push(document.styleSheets[i]);
    }
    return styleSheets;
}

function getStyleSheetByFileName(filename) {
    let styleSheets = getStyleSheets().filter(d => d.href.endsWith(filename));
    if (0 == styleSheets.length) return null;
    return styleSheets[0];
}

function setPreferredColorScheme(mode = "dark") {
    console.log("Changing color theme to " + mode);
    let styleSheet = getStyleSheetByFileName("/css/main.css");
    for (var i = styleSheet.rules.length - 1; i >= 0; i--) {
        rule = styleSheet.rules[i].media;
        if (!rule) continue;
        if (rule.mediaText.includes("prefers-color-scheme")) {
            switch (mode) {
                case "light":
                    //console.log("light");
                    rule.appendMedium("original-prefers-color-scheme");
                    if (rule.mediaText.includes("light")) rule.deleteMedium("(prefers-color-scheme: light)");
                    if (rule.mediaText.includes("dark")) rule.deleteMedium("(prefers-color-scheme: dark)");
                    break;
                case "dark":
                    //console.log("dark");
                    rule.appendMedium("(prefers-color-scheme: light)");
                    rule.appendMedium("(prefers-color-scheme: dark)");
                    if (rule.mediaText.includes("original")) rule.deleteMedium("original-prefers-color-scheme");
                    break;
                default:
                    //console.log("default");
                    rule.appendMedium("(prefers-color-scheme: dark)");
                    if (rule.mediaText.includes("light")) rule.deleteMedium("(prefers-color-scheme: light)");
                    if (rule.mediaText.includes("original")) rule.deleteMedium("original-prefers-color-scheme");
            }
            //console.log(rule);
            //break;
        }
    }
}


setPreferredColorScheme(mode="light");

