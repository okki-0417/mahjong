//画像を取得する
let tiles_img = new Image();
//tiles_img.src = "img/tiles_img.png";
tiles_img.src = "img/tiles_img2.png";

let tiles_in_hand_img = new Image();
tiles_in_hand_img.src = "img/tiles_in_hand_img.png";

class Tiles
{
    constructor(input_y_in_img, input_x_in_img)
    {
        this.col_in_img = input_x_in_img;
        this.row_in_img = input_y_in_img;
        this.w = 47;
        this.h = 63;
        this.x_in_img = this.col_in_img * this.w;
        this.y_in_img = this.row_in_img * this.h;
    }
}

let tiles =
[
    new Tiles(0, 0),		//0.m1
    new Tiles(0, 1),		//1.m2
    new Tiles(0, 2),		//2.m3
    new Tiles(0, 3),		//3.m4
    new Tiles(0, 4),		//4.m5
    new Tiles(0, 5),		//5.m6
    new Tiles(0, 6),		//6.m7
    new Tiles(0, 7),		//7.m8
    new Tiles(0, 8),		//8.m9
    new Tiles(1, 0),		//9.p1
    new Tiles(1, 1),		//10.p2
    new Tiles(1, 2),		//11.p3
    new Tiles(1, 3),		//12.p4
    new Tiles(1, 4),		//13.p5
    new Tiles(1, 5),		//14.p6
    new Tiles(1, 6),		//15.p7
    new Tiles(1, 7),		//16.p8
    new Tiles(1, 8),		//17.p9
    new Tiles(2, 0),		//18.s1
    new Tiles(2, 1),		//19.s2
    new Tiles(2, 2),		//20.s3
    new Tiles(2, 3),		//21.s4
    new Tiles(2, 4),		//22.s5
    new Tiles(2, 5),		//23.s6
    new Tiles(2, 6),		//24.s7
    new Tiles(2, 7),		//25.s8
    new Tiles(2, 8),		//26.s9
    new Tiles(3, 0),		//27.t
    new Tiles(3, 1),		//28.n
    new Tiles(3, 2),		//29.sh
    new Tiles(3, 3),		//30.p
    new Tiles(3, 4),		//31.hk
    new Tiles(3, 5),		//32.ht
    new Tiles(3, 6)		    //33.ch
]