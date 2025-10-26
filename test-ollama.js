// Test script for Ollama functionality
const axios = require('axios');

async function testOllama() {
    console.log('🧠 Testing Ollama integration...\n');
    
    try {
        // Test 1: Check Ollama health
        console.log('1. Checking Ollama health...');
        const healthResponse = await axios.get('http://localhost:11434/api/tags');
        console.log('✅ Ollama is running');
        console.log('📋 Available models:', healthResponse.data.models?.map(m => m.name) || 'None');
        
        // Test 2: Check if our model is available
        const models = healthResponse.data.models || [];
        const ourModel = 'qwen2.5:0.5b';
        const modelExists = models.some(m => m.name === ourModel);
        
        if (modelExists) {
            console.log(`✅ Model ${ourModel} is available`);
        } else {
            console.log(`❌ Model ${ourModel} is not available`);
            console.log('💡 You may need to pull the model: ollama pull qwen2.5:0.5b');
        }
        
        // Test 3: Test basic generation
        console.log('\n2. Testing basic text generation...');
        try {
            const generateResponse = await axios.post('http://localhost:11434/api/generate', {
                model: ourModel,
                prompt: 'Привет! Как дела?',
                stream: false
            });
            
            console.log('✅ Text generation works');
            console.log('🤖 Response:', generateResponse.data.response);
        } catch (error) {
            console.log('❌ Text generation failed:', error.message);
        }
        
        // Test 4: Test memory/context
        console.log('\n3. Testing memory and context...');
        try {
            // First message
            const msg1 = await axios.post('http://localhost:11434/api/generate', {
                model: ourModel,
                prompt: 'Меня зовут Алексей. Запомни это.',
                stream: false
            });
            console.log('📝 First message:', msg1.data.response);
            
            // Second message (should remember the name)
            const msg2 = await axios.post('http://localhost:11434/api/generate', {
                model: ourModel,
                prompt: 'Как меня зовут?',
                stream: false
            });
            console.log('🧠 Memory test:', msg2.data.response);
            
            if (msg2.data.response.toLowerCase().includes('алексей')) {
                console.log('✅ Ollama has memory capabilities');
            } else {
                console.log('❌ Ollama may not have persistent memory');
            }
            
        } catch (error) {
            console.log('❌ Memory test failed:', error.message);
        }
        
        // Test 5: Test financial analysis
        console.log('\n4. Testing financial analysis...');
        try {
            const financialPrompt = `Проанализируй финансовые данные и дай краткие рекомендации на русском языке:

Период: день
Расходы: 1500.00 руб
Доходы: 5000.00 руб
Баланс: 3500.00 руб

Изменения:
- Расходы: 10.0% (📈)
- Доходы: 5.0% (📈)
- Баланс: 15.0% (📈)

Аномалии: 0
Тренды: 1

Дай 2-3 кратких совета по управлению финансами. Будь позитивным и мотивирующим.`;

            const financialResponse = await axios.post('http://localhost:11434/api/generate', {
                model: ourModel,
                prompt: financialPrompt,
                stream: false
            });
            
            console.log('✅ Financial analysis works');
            console.log('💰 AI Response:', financialResponse.data.response);
            
        } catch (error) {
            console.log('❌ Financial analysis failed:', error.message);
        }
        
    } catch (error) {
        console.log('❌ Ollama is not running or not accessible');
        console.log('💡 Make sure Ollama is running: docker-compose up ollama');
        console.log('Error:', error.message);
    }
}

testOllama();
