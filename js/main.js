/*################################################
        グローバル変数等宣言
##################################################*/
//キャンバスの準備
let can = document.getElementById("can");
let con = can.getContext("2d");
const CANVAS_W = 1000;
const CANVAS_H = 600;
can.width = CANVAS_W
can.height = CANVAS_H;
can.style.border = "4px solid";

let player_tiles = [];
let cpu_tiles = [[],[],[]];
let deck = [];
let player_tsumo_tile = [];
let player_throw_away_tile = [];
let cpu_tsumo_tile = [];
let cpu_throw_away_tile = [[],[],[]];
let players_thrown_away_tiles = [[],[],[],[]];
let finish = false;
//let is_win = false;

//王牌の数
const THE_NUMBER_OF_TILES_IN_WAN_PAI = 14;
//数牌の数字(一から)
let number_tiles_number;
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
let  tiles_init_number;

//字牌に色は無いのでz(字)で統一
const CHARACTER_TILES_COLOR = "z";
//字牌の東南西北白發中を設定する
const CHARACTER_TILES_TYPES = 
    ["t","n","sh","p","hk","hts","ch"];
    
//牌の個数(可変)
let the_number_of_tiles_in_hand;

//自家手牌配列の中での理牌時の優先度キーの要素番号
const INIT_NUMBER = 2;
//取得対象画像の取得先チップ画像内での座標
let tile_x_in_img;
let tile_y_in_img;
//取得先のチップ画像内での牌の画像の大きさを取得する
const tile_w = tiles[0].w;
const tile_h = tiles[0].h;

//CPUの数
const THE_NUMBER_OF_CPU = 3;

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

//カーソルが選択している牌配列の要素番号
let cursor_selected_tile = 0;

//捨て牌を選ぶ際の疑似待ち時間のループ
let is_waiting_player_selects_throw_away_tile = false;

//CPUの行動遅延のループ
let is_wait = false;
let cpu_n = 0;

//エレベーターのサイズ
const elevator_size = 150;

//河の一行に置ける牌の数
const TILES_IN_RIVER = 6;

//牌を取得した大きさから縮小するための値
const shrink = 0.7;

//他家が牌を捨てる時の遅延
const WAIT_TIME = 300;

/*#################################################
        処理
###################################################*/
//画像などすべてのロードを終えてからプログラム開始
window.onload = function(){
    start();
}


/*#################################################
        関数
###################################################*/

//メインの関数
function start(){
    prepare_game();

    //次の処理
    player_action();
}

function clear_game(){
    /*deck.splice(0,deck.length);
    player_tiles.splice(0,player_tiles.length);

    for(let i=0; i<THE_NUMBER_OF_CPU; i++){
        cpu_tiles.splice(0,cpu_tiles.length);
    }

    for(let i=0; i<players_thrown_away_tiles.length; i++){
        players_thrown_away_tiles[i].splice(0,players_thrown_away_tiles[i].length);
    }*/
    
    player_tiles = [];
    cpu_tiles = [[],[],[]];
    deck = [];
    player_tsumo_tile = [];
    player_throw_away_tile = [];
    cpu_tsumo_tile = [];
    cpu_throw_away_tile = [[],[],[]];
    players_thrown_away_tiles = [[],[],[],[]];
}

/*---------------------------------------------------
    ゲームの準備関連
----------------------------------------------------*/
//ゲームの全準備。自家の手牌、他家の手牌、山を設定する
function prepare_game(){
    prepare_deck();
    console.log(deck);
    prepare_players_tiles();
    draw_all();
}


//デッキの準備。全ての牌を作る
function prepare_deck(){
    number_tiles_number = 1;
    tiles_init_number = 0;
    /*概要：山の情報の設定
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
    /*概要：自家と他家の手牌情報の設定
        山から13枚引く*/

    //自家が山から13枚引く
    for(let i=0; i<THE_NUMBER_OF_TILES_IN_HAND; i++){
        player_tiles.push(deck.pop());
    }
    riipai(player_tiles);

    //他家が山から13枚引く×3回
    for(let i=0;i<THE_NUMBER_OF_CPU;i++){
        for(let j=0; j<THE_NUMBER_OF_TILES_IN_HAND; j++){
            cpu_tiles[i].push(deck.pop());
        }
        riipai(cpu_tiles[i]);
    }
}

//理牌する
function riipai(tiles){
    //手牌の優先度キーを参照してバブルソートで昇順に並び替える
    bubble_sort(tiles);
}

//バブルソート
function bubble_sort(list){
    for(let i=0; i<list.length; i++){
        for(let j=list.length-1; j>i; j--){
            if(list[j][INIT_NUMBER] < list[j-1][INIT_NUMBER]){
                tmp = list[j];
                list[j] = list[j-1];
                list[j-1] = tmp;
            }
        }
    }
}


/*---------------------------------------------------
    自家のアクション関連
----------------------------------------------------*/
//自家の全てのアクション
function player_action(){
    //残り牌数が王牌以上だったら行動する
    if(deck.length > THE_NUMBER_OF_TILES_IN_WAN_PAI){
        //ツモる        
        tsumo_action(player_tsumo_tile, player_tiles);
        //プレイヤーが捨てる牌を選んで捨てる
        player_throw_away_action();
    }
    else{
        finish = true;
        console.log("山が無くなりました！");
        return;
    }
}


//手牌配列に山から1枚加えるアクション
function tsumo_action(tsumo_array, tiles){
    //ツモ牌を山から1枚引く
    tsumo_array = deck.pop();
    //ツモ牌を手牌に加える
    tiles.push(tsumo_array);
    //新しい情報を描画する
    draw_all();
}

//自家の打牌アクション
function player_throw_away_action(){
    //カーソルを描画する
    draw_cursor();
    //打牌選択ループに入るフラグをtureにする
    is_waiting_player_selects_throw_away_tile = true;
    //打牌選択ループに入る
    throw_away_loop();
}

//打牌選択を待つループ処理
function throw_away_loop(){
    if (is_waiting_player_selects_throw_away_tile){

        //下記のキーボード操作でループフラグがfalseにされない限り1秒に10回(遅延が気にならない適当な時間)ループし続ける
        //(謎)なぜかダブルクオーテーションをつけると無限ループがエラー扱いにならず止まらずに処理できたからこうしてる
        setTimeout("throw_away_loop()", 100);
    }
    else{
        //プレイヤー達全員の捨て牌を格納する配列に捨て牌を格納する
        //配列は2次元で中に自家→下家→対面→上家の順で配列を用意した
        players_thrown_away_tiles[0].push(player_tiles[cursor_selected_tile]);
        //プレイヤーの手牌から該当の牌を削除する
        player_tiles.splice(cursor_selected_tile,1);
        //再度理牌して並び替える
        riipai(player_tiles);

        //新しい情報を描画し直す
        draw_all();

        //次の処理
        judge_tempai();
    }
}

//キーボードイベント
document.onkeydown = function(e){
    //自家が捨て牌を選ぶ時間じゃない時＝自家の捨て牌選択ループ外では動かない
    switch(e.keyCode){
        case 37:    //左
            if (is_waiting_player_selects_throw_away_tile){
                //カーソルが一番左じゃない限り左にカーソルを移動させる
                if(cursor_selected_tile != most_left_tile ){
                    cursor_selected_tile--;

                    //新しいカーソルの位置を描画し直す
                    draw_all();
                    draw_cursor();
                }
            }
            break;
        case 39:    //右
            if (is_waiting_player_selects_throw_away_tile){
                //ツモ時とそれ以外で手牌の数が変わるので選択できる手牌の数を宣言し直す
                most_right_tile = player_tiles.length - 1;
                if(cursor_selected_tile != most_right_tile ){
                    cursor_selected_tile++;

                    draw_all();
                    draw_cursor();
                }
            }
            break;
        //エンターを押したら捨て牌選択待ちのループを解除して次の処理に行けるようにする
        //つまり、エンターで捨て牌を決定して捨てるということ
        case 13:    //エンター
            if (is_waiting_player_selects_throw_away_tile)
                is_waiting_player_selects_throw_away_tile = false;
            break;
        case 82:
            if(finish){
                clear_game();
                start();
            }
    }
}


//自家がテンパイしているか判定する
function judge_tempai(){
    //ToDo テンパイ判定(もシャンテン数判定も)プログラムが難しすぎて挫折
    //次の処理
    is_wait = true;
    wait_cpu_loop();
}


/*---------------------------------------------------
    他家のアクション関連
----------------------------------------------------*/
//他家がツモって捨てる(ツモ切り)
function cpu_actions(i){
    //ツモできる牌がある(残り牌が王牌の数より多い)時だけ動く
    if(deck.length > THE_NUMBER_OF_TILES_IN_WAN_PAI){
        tsumo_action(cpu_tsumo_tile, cpu_tiles[i]);

        //ToDo とりあえず手牌の中から配列番号をランダムに選んで捨ててる
        let random = Math.floor(Math.random() * player_tiles.length);
        //cpu_throw_away_tile[i].push(cpu_tiles[i][random]);
        console.log(cpu_tiles[i][random]);
        players_thrown_away_tiles[i+1].push(cpu_tiles[i][random]);
        cpu_tiles[i].splice(random, 1);

        //捨ててから理牌
        //してもそんなに意味ないけど
        riipai(cpu_tiles[i]);

        //描画し直し
        draw_all();

        if(cpu_n == 2){
            cpu_n = 0;
            player_action();
        }
        else{
            cpu_n++;
            is_wait = true;
            wait_cpu_loop();
        }  
    }
    else{
        console.log("山が無くなりました！")
        finish = true;
    }        
}


//
function wait_cpu_loop(){
    if(is_wait){
        is_wait = false;
        setTimeout(wait_cpu_loop, WAIT_TIME);
    }
    else{
        cpu_actions(cpu_n);
    }
}



/*---------------------------------------------------
    描画関連
----------------------------------------------------*/
//全てを描画
function draw_all(){
    clear_canvas();
    draw_tiles();
    draw_cpu_hands();
    draw_elevator();
    draw_tiles_count();
    draw_thrown_away_tiles();
}

//手牌を描画
function draw_tiles(){
    //手牌の数を再宣言
    the_number_of_tiles_in_hand = player_tiles.length;
    //牌を描画する初期値を設定する
    draw_x = (can.width / 2) - (tile_w * (the_number_of_tiles_in_hand / 2));
    draw_y = can.height - tile_h;

    //手牌13枚の牌配列の優先度キーから画像を取得し描画する
    for(let i=0; i < the_number_of_tiles_in_hand; i++){
        //自家手牌の優先度キーから、対象の牌の画像が取得先のチップ画像内でどの座標にあるのかを得る
        tile_x_in_img = tiles[player_tiles[i][INIT_NUMBER]].x_in_img;
        tile_y_in_img = tiles[player_tiles[i][INIT_NUMBER]].y_in_img;

        //描画する
        con.drawImage( tiles_img, tile_x_in_img, tile_y_in_img, tile_w, tile_h,
            draw_x, draw_y, tile_w, tile_h);
        
        //次の描画位置を設定してあげる
        draw_x += tile_w;
    }
    //カーソル描画の際にこれらの座標を使うのでバグが起こらないよう初期値に戻しておく
    draw_x = (can.width / 2) - (tile_w * ( the_number_of_tiles_in_hand / 2));
    draw_y = can.height - tile_h;
}


//捨て牌を選ぶカーソルを描画
function draw_cursor(){
    //カーソルの位置を指定
    cursor_x = draw_x + ((tile_w - cursor_w)/2) + (cursor_selected_tile * tile_w);
    cursor_y = draw_y - (cursor_h + 3);

    //とりあえず緑の背景で目立つ黄色で描画
    con.fillStyle = "yellow";
    con.fillRect(cursor_x, cursor_y, cursor_w, cursor_h);
}

//キャンバスをクリアする
function clear_canvas(){
    con.clearRect(0,0,can.width,can.height);
}

//CPUの手牌を表示する(伏せているので実際の手牌は反映していない)
function draw_cpu_hands(){
    the_number_of_tiles_in_hand = cpu_tiles[0].length;
    //cpuの手牌画像の取得先での座標と、画像の大きさを宣言
    //縦向きか横向きかで取得すべき座標と大きさが違うので配列で指定
    //それぞれ下家→対面→上家の情報を格納している
    const tile_x_in_img = [0,44,0];
    const tile_y_in_img = [0, 0, 0];
    const tile_w = [44, 33, 44];
    const tile_h = [49, 59, 49];
    //上家と下家の手牌描画時の重なりを考慮した値
    const tile_h_adjusted = 32
    
    //描画の初期位置を設定。各場所の中心に手牌が来るように設定
    let draw_x = [can.width-tile_w[0],
                    (can.width/2) - (tile_w[1] * (the_number_of_tiles_in_hand/2)),
                    0];
    let draw_y = [(can.height/2) - (tile_h_adjusted * (the_number_of_tiles_in_hand/2)),
                    0,
                    (can.height/2) - (tile_h_adjusted * (the_number_of_tiles_in_hand/2))];

    //3人分手牌を設定した位置に描画
    //人数分、手牌の数分描画する
    for(let i=0;i<THE_NUMBER_OF_CPU;i++){
        //手牌の長さを再宣言
        the_number_of_tiles_in_hand = cpu_tiles[i].length;

        for(let j=0;j<the_number_of_tiles_in_hand;j++){
            //描画
            con.drawImage(tiles_in_hand_img, tile_x_in_img[i], tile_y_in_img[i], tile_w[i], tile_h[i],
                        draw_x[i], draw_y[i], tile_w[i], tile_h[i]);
            
            //次の描画位置を設定
            if(i==0 || i==2){
                draw_y[i] += tile_h_adjusted;
            }
            else{
                draw_x[i] += tile_w[i];
            }
        }
    }
}

//エレベーターを描画する
function draw_elevator(){
    con.fillStyle = "black";
    con.fillRect( can.width/2-(elevator_size/2), can.height/2-(elevator_size/2), elevator_size, elevator_size);
}

//山の残り枚数を表示する
function draw_tiles_count(){
    let text = "残り：" + deck.length + "枚";
    let text_w = 120;
    con.font = "20px Arial";
    con.fillStyle = "white";
    con.fillText(text, can.width/2 - text_w/2, can.height/2);
}

//捨てられた牌を河に描画する
function draw_thrown_away_tiles(){
    //捨て牌が無かったら返す
    if(!players_thrown_away_tiles[0].length)return;

    //取得したい画像の取得先での座標と大きさを宣言しておく
    let tile_x_in_img = [];
    let tile_y_in_img = [];
    //取得先の画像サイズ
    //牌の向きによって取得したい画像が違うので大きさを配列に格納しておく
    //順に自家→下家→対面→上家の取得画像の大きさを宣言している
    let tile_w_in_img = [47, 63, 47, 63];
    let tile_h_in_img = [63, 47, 63, 47];

    //描画する画像サイズ
    //取得先の画像サイズに調整する変数をかける
    let tile_w = [];
    let tile_h = [];
    for(let i=0; i<4; i++){
        tile_w[i] = Math.floor(tile_w_in_img[i] * shrink);
        tile_h[i] = Math.floor(tile_h_in_img[i] * shrink);
    }

    //描画位置の初期設定
    const space_between_elevator_and_river = 25;
    const draw_x = [(can.width/2) - (tile_w[0] * (TILES_IN_RIVER/2)),
                    (can.width/2) + elevator_size/2 + space_between_elevator_and_river,
                    (can.width/2) + (tile_w[2] * (TILES_IN_RIVER/2)) - tile_w[2],
                    (can.width/2) - elevator_size/2 - tile_w[3] - space_between_elevator_and_river];

    let draw_y = [(can.height/2) + (elevator_size/2) + space_between_elevator_and_river,
                    (can.height/2) + tile_h[1] * (TILES_IN_RIVER/2) - tile_h[1],
                    (can.height/2) - (elevator_size/2) - tile_h[2] - space_between_elevator_and_river,
                    (can.height/2) - tile_h[3] * (TILES_IN_RIVER/2)];

    let river_row;

    //人数分、捨て牌の数の分描画する
    for(let i=0; i<THE_NUMBER_OF_CPU + 1; i++){
        //捨て牌が無かったら終わる
        if(!players_thrown_away_tiles[i].length)break;
        river_row = 0;

        for(let j=0; j<players_thrown_away_tiles[i].length; j++){
            //捨て牌の画像の取得先の座標を、捨て牌配列内の牌の優先度キーから取得する
            tile_x_in_img[i] = tiles[players_thrown_away_tiles[i][j][INIT_NUMBER]].col_in_img * tile_w_in_img[i];
            //優先度キーと牌の種類の関係は崩したくないが、同じ種類で向きが違う牌が出てくるとその関係維持が難しい
            //なので苦肉の策として末尾の(i*tile_h[0]*4)(←縦のサイズ63であれば[0]じゃなくても2でもいい)で優先度キーと牌の繋がりはそのままに取得する画像の座標を調整した
            //ToDo Tilesクラスで定数this.tile_w/hを使ってthis.x/y_in_imgを宣言してしまっていて大きさが違う画像の座標が指定しにくくなっているので直す
            tile_y_in_img[i] = tiles[players_thrown_away_tiles[i][j][INIT_NUMBER]].row_in_img * tile_h_in_img[i] + (i * tile_h_in_img[0] * 4);

            if(i == 0){
                //もし6牌以上捨て牌がある場合はその分改行して描画する
                if(Math.floor(j / TILES_IN_RIVER) - river_row){
                    draw_y[i] += tile_h[i];
                    draw_x[i] = (can.width/2) - (tile_w[0] * (TILES_IN_RIVER/2));
                    river_row++;
                }

                //描画
                con.drawImage(tiles_img, tile_x_in_img[i], tile_y_in_img[i], tile_w_in_img[i], tile_h_in_img[i],
                    draw_x[i], draw_y[i], tile_w[i], tile_h[i]);
                
                //次の描画位置を設定
                draw_x[i] += tile_w[i];
            }
            else if(i == 1){
                //もし6牌以上捨て牌がある場合はその分改行して描画する
                if(Math.floor(j / TILES_IN_RIVER) - river_row){
                    draw_x[i] += tile_w[i];
                    draw_y[i] = (can.height/2) + tile_h[1] * (TILES_IN_RIVER/2) - tile_h[1];
                    river_row++;
                }

                //描画
                con.drawImage(tiles_img, tile_x_in_img[i], tile_y_in_img[i], tile_w_in_img[i], tile_h_in_img[i],
                    draw_x[i], draw_y[i], tile_w[i], tile_h[i]);

                //次の描画位置を設定
                draw_y[i] -= tile_h[i];
            }
            else if(i==2){
                //もし6牌以上捨て牌がある場合はその分改行して描画する
                if(Math.floor(j / TILES_IN_RIVER) - river_row){
                    draw_y[i] -= tile_h[i];
                    draw_x[i] = (can.width/2) + (tile_w[2] * (TILES_IN_RIVER/2)) - tile_w[2];
                    river_row++;
                }

                //描画
                con.drawImage(tiles_img, tile_x_in_img[i], tile_y_in_img[i], tile_w_in_img[i], tile_h_in_img[i],
                    draw_x[i], draw_y[i], tile_w[i], tile_h[i]);

                //次の描画位置を設定
                draw_x[i] -= tile_w[i];            
            }
            else{
                //もし6牌以上捨て牌がある場合はその分改行して描画する
                if(Math.floor(j / TILES_IN_RIVER) - river_row){
                    draw_x[i] -= tile_w[i];
                    draw_y[i] = (can.height/2) - tile_h[3] * (TILES_IN_RIVER/2);
                    river_row++;
                }

                //描画
                con.drawImage(tiles_img, tile_x_in_img[i], tile_y_in_img[i], tile_w_in_img[i], tile_h_in_img[i],
                    draw_x[i], draw_y[i], tile_w[i], tile_h[i]);

                //次の描画位置を設定
                draw_y[i] += tile_h[i];
            }
        } 
    }
}

/*
function draw_a_tile_in_river(position, j, size1, size2){
    //河に6個並んだら改行する
    position += Math.floor(j/TILES_IN_RIVER) * size1 * shrink;
    //描画
    con.drawImage(tiles_img, tile_x_in_img[i], tile_y_in_img[i], tile_w[i], tile_h[i],
        draw_x[i], draw_y[i], Math.floor(tile_w[i]*0.7), Math.floor(tile_h[i]*0.7));

    console.log(tile_x_in_img, tile_y_in_img, tile_w[i], tile_h[i],
        draw_x[i], draw_y[i], Math.floor(tile_w[i]*0.7), Math.floor(tile_h[i]*0.7));

    //次の描画位置を設定
    draw_x[i] += tile_w[i] * shrink;
}*/






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