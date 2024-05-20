/*-------------------------------------------------
        変数等宣言
-------------------------------------------------*/
//キャンバスの準備
let can = document.getElementById("can");
let con = can.getContext("2d");
const CANVAS_W = 1000;
const CANVAS_H = 520;
can.width = CANVAS_W;
can.height = CANVAS_H;
can.style.border = "4px solid";

let player_tiles = [];
let cpu_tiles = [[],[],[]];
let deck = [];
let player_tsumo_tile = [];
let cpu_throwaway_tile = [];
let is_win = false;

//数牌の数字(一から)
let number_tiles_number = 1;
//数牌の数字(九まで)
const MAX_NUMBER_OF_NUMBER_TILES = 9;
//同じ牌の個数
const THE_NUMBER_OF_SAME_TILES = 4;
//ひとつの色の数牌の個数
const THE_NUMBER_OF_NUMBER_TILES_IN_ONE_COLOR = 36;
//手牌の初期個数
const THE_NUMBER_OF_TILES_IN_HAND = 13;
//数牌のマンズ、ピンズ、ソーズの色を設定する
const NUMBER_TILES_COLORS = ["m","p","s"];
//理牌する時用に牌の並びの優先度も決めておく
//あとで優先度をキーとして配列を参照するので0からの方が都合が良い
let  tiles_init_number = 0;

//字牌に色は無いのでz(字)で統一
const CHARACTER_TILES_COLOR = "z";
//字牌の東南西北白發中を設定する
const CHARACTER_TILES_TYPES = 
    ["t","n","sh","p","hk","hts","ch"];
    
//牌の個数(可変)
let the_number_of_tiles_in_hand;

//自家手牌配列の中での優先度キーの要素番号
const INIT_NUMBER = 2;
//取得対象画像の取得先チップ画像内での座標
let tile_x_in_img;
let tile_y_in_img;
//取得先のチップ画像内での牌の画像の大きさを取得する
const tile_w = tiles[0].w;
const tile_h = tiles[0].h;
//牌を描画する際の描画先の初期座標
let draw_x;
let draw_y;

//捨てる牌を選ぶ待ち時間のループを定義
let interval_id;
let when_select_tile = false;
const LOOP_PACE = 100;

//牌を選択するカーソルの大きさ・場所を定義
const cursor_w = 5;
const cursor_h = 15;
let cursor_x;
let cursor_y;

//カーソルが移動できる一番左、右の牌(配列の一番左、右)を定義
const most_left_tile = 0;
let most_right_tile;

//カーソルが選択している牌
let cursor_selected_tile = 0;


/*-------------------------------------------------
        処理
-------------------------------------------------*/
//画像などすべてのロードを終えてからプログラム開始
window.onload = function(){
    prepare_game();
    player_action();
    //main();
}

//キーボードの操作に応じて呼び出す
document.onkeydown = function(e){
    switch(e.keyCode){
        case 37:    //左
            if(cursor_selected_tile != most_left_tile ){
                cursor_selected_tile--;
                draw_cursor();
            }
            break;
        case 39:    //右
            most_right_tile = player_tiles.length - 1;
            if(cursor_selected_tile != most_right_tile ){
                cursor_selected_tile++;
                draw_cursor();
            }
            break;
        case 13:    //エンター
            if(when_select_tile){
                when_select_tile = false;
            }
    }
}

/*-------------------------------------------------
        関数
-------------------------------------------------*/

function main(){
    prepare_game();

    while(true){
        player_action();
        if(judge_tempai()) break;

        for(let i=0;i<3;i++){
            cpu_action();
        }
        //break;
    }
    riich();
    while(true){
        for(let i=0;i<0;i++){
            cpu_action();
            if(judge_ron()){
                is_win = true;
                break;
            }
        }
        if(is_win) break;

        if(player_in_riich_tusmo()){
            is_win = true;
            break;
        }
        //break;
    }

    console.log("和了！")
}


//ゲームの準備。自家の手牌、他家の手牌、山を設定する
function prepare_game(){
    prepare_deck();
    prepare_players_tiles();
    riipai();
    draw_tiles();
}


//デッキの準備。全ての牌を作る
function prepare_deck(){
    /*出力
    ：山の情報の設定
        数牌：[牌の色、牌の数字、優先度],[〃, 〃, 〃],...
        字牌：[牌の色(字)、牌の種類、優先度],[〃, 〃, 〃],...
        これらを配列deckに入れ、シャッフルする*/

    //数牌の牌を作る
    //各色ごとに一から九の牌を4つずつ作る
    NUMBER_TILES_COLORS.forEach(function(color){
        for(let i=0; i<THE_NUMBER_OF_NUMBER_TILES_IN_ONE_COLOR; i++){   
            //同じ数・優先度の牌を4つずつ作れるように、
            //数字と優先度が4回に1回だけ増えるようにする
            if(i!=0 && i % THE_NUMBER_OF_SAME_TILES == 0){
                number_tiles_number++;
                tiles_init_number++;
            }
            
            //数牌の数字が九を超えたら一に戻す
            if(number_tiles_number > MAX_NUMBER_OF_NUMBER_TILES)
                number_tiles_number = 1;
            
            //山の配列に1牌ごとに[色、数、優先度]の情報を入れる
            deck.push([color, number_tiles_number, tiles_init_number]);
        }
        number_tiles_number++;
        tiles_init_number++;
    });

    

    //字牌の牌を作る
    //字牌の各種類ごとに同じ牌を4つずつ作る
    CHARACTER_TILES_TYPES.forEach(function(type){
        for(let i=0; i< THE_NUMBER_OF_SAME_TILES; i++){
                deck.push([CHARACTER_TILES_COLOR, type, tiles_init_number]);
            }
            number_tiles_number++;
            tiles_init_number++;
    });
    //牌をシャッフルする
    arrayShuffle(deck);
}


//牌をシャッフルする
function arrayShuffle(array){
    for(let i = (array.length - 1); 0 < i; i--){
  
      // 0〜(i+1)の範囲で値を取得
      let r = Math.floor(Math.random() * (i + 1));
  
      // 要素の並び替えを実行
      let tmp = array[i];
      array[i] = array[r];
      array[r] = tmp;
    }
    return array;
}

//自家・他家の手牌を準備
function prepare_players_tiles(){
    /*出力
    ：自家と他家の手牌情報
        山から13枚引く*/

    //自家が山から13枚引く
    for(let i=0; i<THE_NUMBER_OF_TILES_IN_HAND; i++){
        player_tiles.push(deck.pop());
    }

    //CPU(他家)の数
    const THE_NUMBER_OF_CPU = 3;

    //他家が山から13枚引く×3回
    for(let i=0;i<THE_NUMBER_OF_CPU;i++){
        for(let j=0; j<THE_NUMBER_OF_TILES_IN_HAND; j++){
            cpu_tiles[i].push(deck.pop());
        }
    }

}

//理牌する
function riipai(){
    bubble_sort(player_tiles);
}

//バブルソート
function bubble_sort(list){
    for(let i=0; i<list.length; i++){
        for(let j=list.length-1; j>i-1; j--){
            if(list[j] < list[j-1]){
                tmp = list[j];
                list[j] = list[j-1];
                list[j-1] = tmp;
            }
        }
    }
}

//手牌の描画
function draw_tiles(){
    clear_tiles();

    the_number_of_tiles_in_hand = player_tiles.length;
    draw_x = (can.width / 2) - (tile_w * ( the_number_of_tiles_in_hand / 2));
    draw_y = can.height - tile_h;

    //手牌13枚の牌配列の優先度キーから画像を取得し描画する
    for(i=0; i < the_number_of_tiles_in_hand; i++){
        //自家手牌の優先度キーから、対象の牌の画像が取得先のチップ画像内でどの座標にあるのかを得る
        tile_x_in_img = tiles[player_tiles[i][INIT_NUMBER]].x_in_img;
        tile_y_in_img = tiles[player_tiles[i][INIT_NUMBER]].y_in_img;

        //描画する
        con.drawImage( tiles_img, tile_x_in_img, tile_y_in_img, tile_w, tile_h,
            draw_x, draw_y, tile_w, tile_h);
        
        draw_x += tile_w;
    }
    
    //カーソル描画の際に使うので初期値に戻しておく
    draw_x = (can.width / 2) - (tile_w * ( the_number_of_tiles_in_hand / 2));
    draw_y = can.height - tile_h;
}

//手牌のクリア
function clear_tiles(){
    if(draw_x != undefined){
        con.clearRect(0, draw_y, CANVAS_W, tile_h);
    }
}

//自家の全てのアクション
function player_action(){
    //プレイヤーが山から1枚ツモってくる
    player_tsumo_action();
    draw_tiles();
    draw_cursor();
    //プレイヤーが捨てる牌を選んで捨てる
    player_throw_awaw_action();
}

//自家ツモアクション
function player_tsumo_action(){
    //ツモ牌を山から1枚引く
    player_tsumo_tile = deck.pop();
        
    //ツモ牌を手牌に加える
    player_tiles.push(player_tsumo_tile);
}

//捨て牌を選ぶカーソルを描画
function draw_cursor(){
    cleaer_cursor();

    cursor_x = draw_x + ((tile_w - cursor_w)/2) + (cursor_selected_tile * tile_w);
    cursor_y = draw_y - (cursor_h + 3);

    con.fillStyle = "red";
    con.fillRect(cursor_x, cursor_y, cursor_w, cursor_h);
}

//カーソルのクリア
function cleaer_cursor(){
    if(cursor_x != undefined){
        con.clearRect(0, cursor_y, CANVAS_W, cursor_h);
    }
}

//自家打牌アクション
function player_throw_awaw_action(){
    
    when_select_tile = true;

    interval_id = setInterval(function(){
        if(!when_select_tile) clearInterval(interval_id);
        console.log("a");
    },LOOP_PACE);

    //捨てる牌を選んで捨てる
    //let throw_away_tile = window.prompt("捨てる牌を選択してください");
    if(!when_select_tile){
        player_tiles.splice(cursor_selected_tile,1);
        draw_tiles();
        cleaer_cursor();
    }
}

//自家がテンパイしているか判定する
function judge_tempai(){

}

//他家がツモって捨てる(ツモ切り)
function cpu_action(){

}

//リーチの表示をする
function riich(){

}

//リーチ中、他家が捨てた牌がロンか判定する
function judge_ron(){

}

//リーチ中、ツモる
function player_in_riich_tusmo(){

}

//ツモった牌がアガリ牌か判定する
function judge_tsumo(){
    
}