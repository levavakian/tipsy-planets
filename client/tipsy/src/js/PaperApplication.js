
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

        paper.install(window)
        paper.setup(document.getElementById('canvas'));
        
        paper.view.onFrame = function(event){
            path.rotate(3)
        }

        var raster = new paper.Raster(db);
        raster.position = paper.view.center

        paper.view.onResize = function(event) {
            raster.position = paper.view.center
        }

        var path = new paper.Path.Rectangle({
            point: [75, 75],
            size: [75, 75],
            strokeColor: 'black'
        });
    }
}
