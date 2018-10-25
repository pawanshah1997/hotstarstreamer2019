const express = require('express')
const app = express()
const port = 3000
const puppeteer = require('puppeteer');

// Puppeteer page event types to catch
const pevents = [
    'response'
];
app.get('/', async (req, res) => { 
     // Create headless session
     const browser = await puppeteer.launch({ args: ['--no-sandbox', '--disable-setuid-sandbox'] });
     const page = await browser.newPage();
     const client = await page.target().createCDPSession();
     // Log puppeter page notifications
     pevents.forEach((peventName) => {
         page.on(peventName, async (plistenerFunc) => {
             //console.log({ peventName, plistenerFunc });
             if (peventName == 'response') {
                 await plistenerFunc.text()
                     .then((textBody) => {
                         if(textBody.indexOf('{"body":{"results":{"item":{"playbackUrl"')!== -1){
                            res.send(textBody);
                         }
                     }, (err) => {
                         //console.error(plistenerFunc, err);
                         //console.log(plistenerFunc, err);
                     })
                 ;
             }
         });
     });
     // Open a page, than close
     await page.goto(req.query.url, { waitUntil: ['networkidle2', 'load', 'domcontentloaded'], timeout: 100000 });
     await page.close();
     await browser.close();
     res.send({error: 'failed'});
})

app.listen(port, () => console.log(`hotstar parser app listening on port ${port}!`))
