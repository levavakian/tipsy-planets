
import paper from 'paper';
import db from './assets/db.jpg'

export default class PaperApplication {
    constructor() {
        this.version = '1'
        this.blocks = [];
        this.pointLights = [];
        this.init();
    }

    init() {
        console.log('PaperApplication::init');

        let w = window.innerWidth;
        let h = window.innerHeight;

        const canvas = document.createElement('canvas');
        canvas.id = 'paper-canvas';
        document.body.appendChild(canvas);

        paper.setup(canvas);
        
        paper.view.onFrame = function(event){
            path.rotate(3)
        }

        console.log("1")
        console.log(db)
        console.log("1")

        var raster = new paper.Raster(db);
        raster.position = paper.view.center

        var path = new paper.Path.Rectangle({
            point: [75, 75],
            size: [75, 75],
            strokeColor: 'black'
        });
    }
}
